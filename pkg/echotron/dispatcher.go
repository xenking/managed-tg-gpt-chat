/*
 * Echotron
 * Copyright (C) 2018-2022 The Echotron Devs
 *
 * Echotron is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Echotron is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package echotron

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

// SessionHandler is the interface that must be implemented by your definition of
// the struct thus it represent each open session with a user on Telegram.
type SessionHandler interface {
	// Mount will be called upon receiving any update from Telegram.
	HandleUpdate(ctx context.Context, upd *Update)
}

type HandlerFunc func(ctx context.Context, upd *Update)

func (fn HandlerFunc) HandleUpdate(ctx context.Context, upd *Update) {
	fn(ctx, upd)
}

var NoopSessionHandler HandlerFunc = func(ctx context.Context, upd *Update) {}

// NewSessionFactory is called every time echotron receives an update with a chat ID never
// encountered before.
type NewSessionFactory func(chatId int64) SessionHandler

// The Dispatcher passes the updates from the Telegram SessionHandler API to the SessionHandler instance
// associated with each chatID. When a new chat ID is found, the provided function
// of type NewSessionFactory will be called.
type Dispatcher struct {
	sessionMap map[int64]SessionHandler
	newSession NewSessionFactory
	updates    chan *Update
	httpServer *http.Server
	api        API
	mu         sync.RWMutex
}

// NewDispatcher returns a new instance of the Dispatcher object.
// Calls the Update function of the bot associated with each chat ID.
// If a new chat ID is found, newBotFn will be called first.
func NewDispatcher(token string, newBotFn NewSessionFactory) *Dispatcher {
	return &Dispatcher{
		api:        NewAPI(token),
		sessionMap: make(map[int64]SessionHandler),
		newSession: newBotFn,
		updates:    make(chan *Update),
	}
}

// ListenUpdates listens for updates from the Telegram API and calls the HandleUpdate function
func (d *Dispatcher) ListenUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-d.updates:
			bot := d.instance(update.ChatID())
			go bot.HandleUpdate(ctx, update)
		}
	}
}

// DeleteSession deletes the SessionHandler instance, seen as a session, from the
// map with all of them.
func (d *Dispatcher) DeleteSession(chatID int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.sessionMap, chatID)
}

// AddSession allows to arbitrarily create a new SessionHandler instance.
func (d *Dispatcher) AddSession(chatID int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, isIn := d.sessionMap[chatID]; !isIn {
		d.sessionMap[chatID] = d.newSession(chatID)
	}
}

// Poll is a wrapper function for PollOptions.
func (d *Dispatcher) Poll(ctx context.Context) error {
	return d.PollOptions(ctx, true, UpdateOptions{Timeout: 120})
}

// PollOptions starts the polling loop so that the dispatcher calls the function Update
// upon receiving any update from Telegram.
func (d *Dispatcher) PollOptions(ctx context.Context, dropPendingUpdates bool, opts UpdateOptions) error {
	var (
		timeout    = opts.Timeout
		isFirstRun = true
	)

	// deletes webhook if present to run in long polling mode
	if _, err := d.api.DeleteWebhook(ctx, dropPendingUpdates); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if isFirstRun {
			opts.Timeout = 0
		}

		response, err := d.api.GetUpdates(ctx, &opts)
		if err != nil {
			return err
		}

		if !dropPendingUpdates || !isFirstRun {
			for _, u := range response.Result {
				d.updates <- u
			}
		}

		if l := len(response.Result); l > 0 {
			opts.Offset = response.Result[l-1].ID + 1
		}

		if isFirstRun {
			isFirstRun = false
			opts.Timeout = timeout
		}
	}
}

func (d *Dispatcher) instance(chatID int64) SessionHandler {
	d.mu.RLock()
	bot, ok := d.sessionMap[chatID]
	d.mu.RUnlock()
	if !ok {
		bot = d.newSession(chatID)
		d.mu.Lock()
		d.sessionMap[chatID] = bot
		d.mu.Unlock()
	}
	return bot
}

// ListenWebhook is a wrapper function for ListenWebhookOptions.
func (d *Dispatcher) ListenWebhook(ctx context.Context, webhookURL string) error {
	return d.ListenWebhookOptions(ctx, webhookURL, false, nil)
}

// ListenWebhookOptions sets a webhook and listens for incoming updates.
// The webhookUrl should be provided in the following format: '<hostname>:<port>/<path>',
// eg: 'https://example.com:443/bot_token'.
// ListenWebhook will then proceed to communicate the webhook url '<hostname>/<path>' to Telegram
// and run a webserver that listens to ':<port>' and handles the path.
func (d *Dispatcher) ListenWebhookOptions(ctx context.Context, webhookURL string, dropPendingUpdates bool, opts *WebhookOptions) error {
	u, err := url.Parse(webhookURL)
	if err != nil {
		return err
	}

	whURL := fmt.Sprintf("%s%s", u.Hostname(), u.EscapedPath())
	if _, err = d.api.SetWebhook(ctx, whURL, dropPendingUpdates, opts); err != nil {
		return err
	}

	if d.httpServer != nil {
		mux := http.NewServeMux()
		mux.Handle("/", d.httpServer.Handler)
		mux.HandleFunc(u.EscapedPath(), d.HandleWebhook)
		d.httpServer.Handler = mux
		return d.httpServer.ListenAndServe()
	}
	http.HandleFunc(u.EscapedPath(), d.HandleWebhook)
	return http.ListenAndServe(fmt.Sprintf(":%s", u.Port()), nil)
}

// SetHTTPServer allows to set a custom http.Server for ListenWebhook and ListenWebhookOptions.
func (d *Dispatcher) SetHTTPServer(s *http.Server) {
	d.httpServer = s
}

// HandleWebhook is the http.HandlerFunc for the webhook URL.
// Useful if you've already a http server running and want to handle the request yourself.
func (d *Dispatcher) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var update Update

	if err := unmarshalJSONRequest(r, &update); err != nil {
		log.Println("echotron.Dispatcher", "HandleWebhook", err)
		return
	}

	d.updates <- &update
}

func unmarshalJSONRequest(r *http.Request, value interface{}) error {
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			return err
		}
		defer reader.Close()
		return json.NewDecoder(reader).Decode(value)
	default:
		return json.NewDecoder(r.Body).Decode(value)
	}
}
