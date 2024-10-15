// error_handling is a bot that handles zero division panic.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
			tm.IsCommandMessage("start"),
			func(u *tm.Update) {
				msg := tgbotapi.NewMessage(
					u.Message.Chat.ID,
					"Hello! I divide numbers. For example: `/div 20 4`.\n\nHint:  I can handler errors! Try `/div 42 0`",
				)
				msg.ParseMode = "markdown"
				bot.Send(msg)
			},
		)).
		Mount(tm.NewRoute(
			tm.And(tm.IsMessage(), tm.HasRegex(`^/div (\d+) (\d+)$`)),
			func(u *tm.Update) {
				parts := strings.Split(u.Message.Text, " ")
				a, _ := strconv.Atoi(parts[1])
				b, _ := strconv.Atoi(parts[2])
				bot.Send(tgbotapi.NewMessage(
					u.Message.Chat.ID,
					fmt.Sprintf("The result is %d", a/b),
				))
			},
		)).
		SetRecover(func(u *tm.Update, err error, stackTrace string) {
			chat := u.EffectiveChat()
			if chat != nil {
				bot.Send(tgbotapi.NewMessage(
					chat.ID,
					fmt.Sprintf("Oops, an error occurred: %s", err),
				))
				log.Printf("Warning! An error occurred: %s", err)
			}
		})

	for update := range updates {
		mux.Dispatch(bot, update)
	}
}
