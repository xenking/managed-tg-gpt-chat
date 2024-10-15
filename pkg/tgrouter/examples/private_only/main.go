// private_only is a bot that allows you to talk to him only in private chat.
package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	tm "github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	mux := tm.NewRouter().
		Mount(tm.NewRoute(
			tm.And(tm.IsPrivate(), tm.IsCommandMessage("start")),
			func(u *tm.Update) {
				bot.Send(tgbotapi.NewMessage(
					u.Message.Chat.ID,
					"Psst... Don't tell anyone about our private chat! :)",
				))
			},
		)).
		Mount(tm.NewRoute(
			tm.IsCommandMessage("start"),
			func(u *tm.Update) {
				bot.Send(tgbotapi.NewMessage(
					u.Message.Chat.ID,
					"Sorry, I only respond in private chats. Send me a direct message!",
				))
			},
		))

	for update := range updates {
		mux.Dispatch(bot, update)
	}
}
