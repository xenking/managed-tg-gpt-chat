// album_conversation is a bot that allows users to upload & share photos.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
	router "github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

// Photo describes a submitted photo
type Photo struct {
	ID          int
	FileID      string
	Description string
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	botToken := os.Getenv("TG_TOKEN")

	api := echotron.NewAPI(botToken)

	lastID := 0
	var photos []Photo

	handler := router.NewRouter(api).
		Mount(router.NewConversationRoute(
			"upload_photo_dialog",
			router.NewLocalPersistence(), // we could also use `router.NewFilePersistence("db.json")` or `&gormpersistence.GORMPersistence(db)` to keep data across api restarts
			router.StateMap{
				"": router.NewRoute(router.IsCommandMessage("add"), router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
					_, err := u.SendMessage(ctx, "Please send me your photo.", u.Message.Chat.ID, nil)
					if err != nil {
						return err
					}
					u.PersistenceContext.SetState("upload_photo")
					return nil
				})),

				"upload_photo": router.NewGroupAny(
					router.NewRoute(router.HasPhoto(), router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
						data := u.PersistenceContext.GetData()
						data["photoID"] = u.Message.Photo[0].FileID
						u.PersistenceContext.SetData(data)
						_, err := u.SendMessage(ctx, "Please enter photo description.", u.Message.Chat.ID, nil)
						if err != nil {
							return err
						}
						u.PersistenceContext.SetState("enter_description")
						return nil
					})),
					router.NewRoute(router.Not(router.IsCommandMessage("cancel")), router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
						_, err := u.SendMessage(ctx, "Sorry, I only accept photos. Please try again!", u.Message.Chat.ID, nil)
						return err
					})),
				),
				"enter_description": router.NewGroupAny(
					router.NewRoute(router.HasText(), router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
						data := u.PersistenceContext.GetData()
						data["photoDescription"] = u.Message.Text
						u.PersistenceContext.SetData(data)
						opts := &echotron.MessageOptions{
							ReplyMarkup: echotron.ReplyKeyboardMarkup{Keyboard: [][]echotron.KeyboardButton{
								{echotron.KeyboardButton{Text: "Yes"}, echotron.KeyboardButton{Text: "No"}},
							}},
						}
						_, err := u.SendMessage(ctx, "Are you sure you want to save this photo?", u.Message.Chat.ID, opts)
						if err != nil {
							return err
						}
						u.PersistenceContext.SetState("confirm_submission")
						return nil
					})),
					router.NewRoute(router.Not(router.IsCommandMessage("cancel")), router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
						_, err := u.SendMessage(ctx, "Sorry, I did not understand that. Please enter some text!", u.Message.Chat.ID, nil)
						return err
					})),
				),

				"confirm_submission": router.NewRoute(router.And(router.IsCallbackQuery(), router.HasText()), router.HandlerFunc(func(ctx context.Context, u *router.Update) (err error) {
					data := u.PersistenceContext.GetData()
					switch u.Message.Text {
					case "Yes":
						lastID++
						photos = append(photos, Photo{
							lastID,
							data["photoID"].(string),
							data["photoDescription"].(string),
						})
						_, err = u.SendMessage(ctx, "Photo submitted! Type /list to list all photos.", u.Message.Chat.ID, nil)
					case "No":
						_, err = u.SendMessage(ctx, "Cancelled.", u.Message.Chat.ID, nil)
					}
					u.PersistenceContext.ClearData()
					if err != nil {
						return err
					}
					_, err = u.SendMessage(ctx, "Are you sure you want to save this photo?", u.Message.Chat.ID, nil)
					if err != nil {
						return err
					}
					u.PersistenceContext.SetState("")
					return nil
				})),
			},
			router.NewRoute(router.IsCommandMessage("cancel"), router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
				u.PersistenceContext.ClearData()
				u.PersistenceContext.SetState("")
				_, err := u.SendMessage(ctx, "Cancelled.", u.Message.Chat.ID, nil)
				return err
			})),
		)).
		Mount(router.NewRoute(
			router.IsCommandMessage("list"),
			router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
				var lines []string
				for _, photo := range photos {
					lines = append(lines, fmt.Sprintf("- %s (/view_%d)", photo.Description, photo.ID))
				}
				if len(lines) == 0 {
					lines = append(lines, "No photos yet.")
				}
				opts := &echotron.MessageOptions{
					ReplyParameters: echotron.ReplyParameters{
						MessageID: u.Message.ID,
						ChatID:    u.Message.Chat.ID,
					},
				}
				_, err := u.SendMessage(ctx, "Photos:\n"+strings.Join(lines, "\n"), u.Message.Chat.ID, opts)
				return err
			}),
		)).
		Mount(router.NewRoute(
			router.HasRegex(`^/view_(\d+)$`),
			router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
				photoID := strings.Split(u.Message.Text, "_")[1]
				var match *Photo
				for i, photo := range photos {
					if fmt.Sprint(photo.ID) == photoID {
						match = &photos[i]
					}
				}
				if match == nil {
					_, err := u.SendMessage(ctx, "Photo not found!", u.Message.Chat.ID, nil)
					return err
				}
				opts := &echotron.PhotoOptions{
					Caption: fmt.Sprintf("Description: %s", match.Description),
				}
				_, err := u.SendPhoto(ctx, echotron.NewInputFileID(match.FileID), u.Message.Chat.ID, opts)
				return err
			}),
		)).
		Mount(router.NewMessageRoute(
			nil,
			router.HandlerFunc(func(ctx context.Context, u *router.Update) error {
				opts := &echotron.MessageOptions{
					ReplyParameters: echotron.ReplyParameters{
						MessageID: u.Message.ID,
						ChatID:    u.Message.Chat.ID,
					},
				}
				_, err := u.SendMessage(ctx, "Hello! I'm a gallery api.\n\nI allow users to upload & share their photos!\n\nAvailable commands:\n/add - add photo\n/list - list photos", u.Message.Chat.ID, opts)
				return err
			}),
		))

	dsp := echotron.NewDispatcher(botToken, func(chatID int64) echotron.SessionHandler {
		return handler
	})
	go dsp.ListenUpdates(ctx)
	go dsp.Poll(ctx)
}
