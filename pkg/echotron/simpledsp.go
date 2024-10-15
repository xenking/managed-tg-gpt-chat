package echotron

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// PollingUpdates is a wrapper function for PollingUpdatesOptions.
func PollingUpdates(ctx context.Context, token string) <-chan *Update {
	return PollingUpdatesOptions(ctx, token, true, UpdateOptions{Timeout: 120})
}

// PollingUpdatesOptions returns a read-only channel of incoming  updates from the Telegram API.
func PollingUpdatesOptions(ctx context.Context, token string, dropPendingUpdates bool, opts UpdateOptions) <-chan *Update {
	updates := make(chan *Update)

	go func() {
		defer close(updates)

		var (
			api        = NewAPI(token)
			timeout    = opts.Timeout
			isFirstRun = true
		)

		// deletes webhook if present to run in long polling mode
		if _, err := api.DeleteWebhook(ctx, dropPendingUpdates); err != nil {
			log.Println("echotron.PollingUpdates", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if isFirstRun {
				opts.Timeout = 0
			}

			response, err := api.GetUpdates(ctx, &opts)
			if err != nil {
				log.Println("echotron.PollingUpdates", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if !dropPendingUpdates || !isFirstRun {
				for _, u := range response.Result {
					updates <- u
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
	}()

	return updates
}

// WebhookUpdates is a wrapper function for WebhookUpdatesOptions.
func WebhookUpdates(ctx context.Context, url, token string) <-chan *Update {
	return WebhookUpdatesOptions(ctx, url, token, false, nil)
}

// WebhookUpdatesOptions returns a read-only channel of incoming updates from the Telegram API.
// The webhookUrl should be provided in the following format: '<hostname>:<port>/<path>',
// eg: 'https://example.com:443/bot_token'.
// WebhookUpdatesOptions will then proceed to communicate the webhook url '<hostname>/<path>'
// to Telegram and run a webserver that listens to ':<port>' and handles the path.
func WebhookUpdatesOptions(ctx context.Context, whURL, token string, dropPendingUpdates bool, opts *WebhookOptions) <-chan *Update {
	u, err := url.Parse(whURL)
	if err != nil {
		panic(err)
	}

	wURL := u.Hostname() + u.EscapedPath()
	api := NewAPI(token)
	if _, err := api.SetWebhook(ctx, wURL, dropPendingUpdates, opts); err != nil {
		panic(err)
	}

	updates := make(chan *Update)
	http.HandleFunc(u.EscapedPath(), func(w http.ResponseWriter, r *http.Request) {
		var update Update

		if err := unmarshalJSONRequest(r, &update); err != nil {
			log.Println("echotron.WebhookUpdates", err)
			return
		}

		updates <- &update
	})

	go func() {
		defer close(updates)
		port := fmt.Sprintf(":%s", u.Port())
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := http.ListenAndServe(port, nil); err != nil {
				log.Println("echotron.WebhookUpdates", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()

	return updates
}
