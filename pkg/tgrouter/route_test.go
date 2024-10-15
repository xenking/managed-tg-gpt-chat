package tgrouter_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	tm "github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

func ExampleNewCommandHandler() {
	bot, _ := tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	u := tgbotapi.NewUpdate(0)
	updates := bot.GetUpdatesChan(u)
	mux := tm.NewRouter()
	mux.Mount(tm.NewCommandRoute(
		"add",
		nil,
		func(u *tm.Update) {
			args := u.Context["args"].([]string)
			if len(args) != 2 {
				bot.Send(tgbotapi.NewMessage(
					u.EffectiveChat().ID, "Wrong number of arguments. Example: /add 13 37"),
				)
				return
			}
			a, err1 := strconv.Atoi(args[0])
			b, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil {
				bot.Send(tgbotapi.NewMessage(
					u.EffectiveChat().ID, "Arguments must be numbers. Example: /add 13 37"),
				)
				return
			}
			bot.Send(tgbotapi.NewMessage(
				u.EffectiveChat().ID, fmt.Sprintf("%d + %d = %d", a, b, a+b),
			))
		},
	))
	for update := range updates {
		mux.Dispatch(bot, update)
	}
}

func TestHandlerConsume(t *testing.T) {
	a, b, c := false, false, false
	h := tm.NewRoute(
		nil,
		func(u *tm.Update) { a = true },
		func(u *tm.Update) { b = true; u.Consume() },
		func(u *tm.Update) { c = true },
	)
	u := &tm.Update{tgbotapi.Update{}, nil, false, nil, nil}
	if !h.Handle(nil, u) {
		t.Error("routeHandler should return true")
	}
	if !a {
		t.Error("First routeHandler should fire")
	}
	if !b {
		t.Error("Second routeHandler should fire")
	}
	if c {
		t.Error("Third routeHandler should not fire")
	}
}

func TestCommandHandler(t *testing.T) {
	h := tm.NewCommandRoute("test", nil, func(u *tm.Update) {})
	u := &tm.Update{}
	u.Update.Message = &tgbotapi.Message{}
	u.Update.Message.Text = "/test foo bar"
	u.Context = make(map[string]interface{})
	u.API = &tgbotapi.BotAPI{}
	u.API.Self.UserName = "testbot"
	if !h.Handle(nil, u) {
		t.Error("routeHandler should return true")
	}
	args := u.Context["args"].([]string)
	if len(args) != 2 {
		t.Error("There should be 2 args")
	}
	if args[0] != "foo" {
		t.Error("First arg should be 'foo'")
	}
	if args[1] != "bar" {
		t.Error("Second arg should be 'bar'")
	}

	h = tm.NewCommandRoute("foo bar", nil, func(u *tm.Update) {})
	u.Update.Message.Text = "/foo 42"
	if !h.Handle(nil, u) {
		t.Error("routeHandler should process update")
	}
	u.Update.Message.Text = "/bar 42"
	if !h.Handle(nil, u) {
		t.Error("routeHandler should process update")
	}
	u.Update.Message.Text = "/baz 42"
	if h.Handle(nil, u) {
		t.Error("routeHandler should not process update")
	}
}

func TestConversationHandler(t *testing.T) {
	NewUpdate := func(text string) *tm.Update {
		u := &tm.Update{}
		u.Update.Message = &tgbotapi.Message{}
		u.Update.Message.Text = text
		u.Update.Message.From = &tgbotapi.User{}
		u.Update.Message.From.ID = 13
		u.Update.Message.Chat = &tgbotapi.Chat{}
		u.Update.Message.Chat.ID = 37
		u.Context = make(map[string]interface{})
		u.API = &tgbotapi.BotAPI{}
		u.API.Self.UserName = "testbot"
		return u
	}
	askAgeEntered := false
	p := tm.NewLocalPersistence()
	h := tm.NewConversationRoute(
		"test",
		p,
		tm.StateMap{
			"": {
				tm.NewCommandRoute("start", nil, func(u *tm.Update) {
					u.PersistenceContext.SetState("ask_name")
				}),
			},
			"ask_name": {
				tm.NewMessageRoute(tm.HasText(), func(u *tm.Update) {
					data := u.PersistenceContext.GetData()
					data["name"] = u.EffectiveMessage().Text
					u.PersistenceContext.SetData(data)
					u.PersistenceContext.SetState("ask_age")
				}),
			},
			"ask_age:enter": {
				tm.NewRoute(nil, func(u *tm.Update) {
					askAgeEntered = true
				}),
			},
			"ask_age": {
				tm.NewMessageRoute(tm.HasText(), func(u *tm.Update) {
					data := u.PersistenceContext.GetData()
					data["age"] = u.EffectiveMessage().Text
					u.PersistenceContext.SetData(data)
					u.PersistenceContext.SetState("ask_confirm")
				}),
			},
			"ask_confirm": {
				tm.NewCommandRoute("confirm", nil, func(u *tm.Update) {
					u.PersistenceContext.ClearData()
					u.PersistenceContext.SetState("")
				}),
			},
		},
		[]*tm.routeHandler{
			tm.NewCommandRoute("cancel", nil, func(u *tm.Update) {
				u.PersistenceContext.SetState("")
				u.PersistenceContext.ClearData()
			}),
		},
	)
	pk := tm.PersistenceKey{"test", 13, 37}
	assert(!h.Handle(nil, NewUpdate("just some text")), t, "Random text must be ignored")
	assert(h.Handle(nil, NewUpdate("/start")), t, "/start must be processed")
	assert(p.GetState(pk) == "ask_name", t, "State must be ask_name, have", p.GetState(pk))
	assert(!askAgeEntered, t)
	assert(h.Handle(nil, NewUpdate("Foobar")), t, "Name must be processed")
	assert(p.GetState(pk) == "ask_age", t, "State must be ask_age, have", p.GetState(pk))
	assert(askAgeEntered, t)
	assert(reflect.DeepEqual(p.GetData(pk), map[string]interface{}{"name": "Foobar"}), t, "Unexpected persistence data")
	assert(h.Handle(nil, NewUpdate("18")), t, "Age must be processed")
	assert(p.GetState(pk) == "ask_confirm", t, "State must be ask_confirm, have", p.GetState(pk))
	assert(reflect.DeepEqual(p.GetData(pk), map[string]interface{}{"name": "Foobar", "age": "18"}), t, "Unexpected persistence data")
	assert(!h.Handle(nil, NewUpdate("foobar")), t, "Random text must be ignored")
	assert(p.GetState(pk) == "ask_confirm", t, "State must be ask_confirm, have", p.GetState(pk))
	assert(h.Handle(nil, NewUpdate("/confirm")), t, "/confirm must be processed")
	assert(p.GetState(pk) == "", t, "State must be empty, have", p.GetState(pk))
	assert(reflect.DeepEqual(p.GetData(pk), map[string]interface{}{}), t, "Persistence data must be empty")

	assert(h.Handle(nil, NewUpdate("/start")), t, "/start must be processed")
	assert(h.Handle(nil, NewUpdate("OtherUser")), t, "Name must be processed")
	assert(p.GetState(pk) == "ask_age", t, "State must be ask_age, have", p.GetState(pk))
	assert(reflect.DeepEqual(p.GetData(pk), map[string]interface{}{"name": "OtherUser"}), t, "Unexpected persistence data")
	assert(h.Handle(nil, NewUpdate("/cancel")), t, "/cancel must be processed")
	assert(p.GetState(pk) == "", t, "State must be empty, have", p.GetState(pk))
	assert(reflect.DeepEqual(p.GetData(pk), map[string]interface{}{}), t, "Persistence data must be empty")
}

func TestConvenienceHandlers(t *testing.T) {
	assert(strings.HasSuffix(
		getFunctionName(tm.NewInlineQueryRoute(".*", nil, func(u *tm.Update) {}).filter),
		"And.func1",
	), t)
	assert(strings.HasSuffix(
		getFunctionName(tm.NewInlineQueryRoute(".*", tm.Any(), func(u *tm.Update) {}).filter),
		"And.func1",
	), t)
	update := &tm.Update{
		Update: tgbotapi.Update{
			InlineQuery: &tgbotapi.InlineQuery{
				Query: "foo:bar:42",
			},
		},
		Context: map[string]interface{}{},
	}
	tm.NewInlineQueryRoute(`^foo:(\w+):(\d+)`, nil, func(u *tm.Update) {
	}).Handle(nil, update)
	assert(reflect.DeepEqual(update.Context["matches"], []string{"foo:bar:42", "bar", "42"}), t)

	assert(strings.HasSuffix(
		getFunctionName(tm.NewCallbackQueryHandler(".*", nil, func(u *tm.Update) {}).filter),
		"And.func1",
	), t)
	assert(strings.HasSuffix(
		getFunctionName(tm.NewCallbackQueryHandler(".*", tm.Any(), func(u *tm.Update) {}).filter),
		"And.func1",
	), t)
	update = &tm.Update{
		Update: tgbotapi.Update{
			CallbackQuery: &tgbotapi.CallbackQuery{
				Data: "foo:bar:42",
			},
		},
		Context: map[string]interface{}{},
	}
	tm.NewCallbackQueryHandler(`^foo:(\w+):(\d+)`, nil, func(u *tm.Update) {
	}).Handle(nil, update)
	assert(reflect.DeepEqual(update.Context["matches"], []string{"foo:bar:42", "bar", "42"}), t)

	assert(strings.HasSuffix(
		getFunctionName(tm.NewEditedMessageRoute(nil, func(u *tm.Update) {}).filter),
		"NewEditedMessageRoute.func1",
	), t)
	assert(strings.HasSuffix(
		getFunctionName(tm.NewEditedMessageRoute(tm.Any(), func(u *tm.Update) {}).filter),
		"And.func1",
	), t)

	assert(strings.HasSuffix(
		getFunctionName(tm.NewChannelPostRoute(nil, func(u *tm.Update) {}).filter),
		"NewChannelPostRoute.func1",
	), t)
	assert(strings.HasSuffix(
		getFunctionName(tm.NewChannelPostRoute(tm.Any(), func(u *tm.Update) {}).filter),
		"And.func1",
	), t)

	assert(strings.HasSuffix(
		getFunctionName(tm.NewEditedChannelPostRoute(nil, func(u *tm.Update) {}).filter),
		"NewEditedChannelPostRoute.func1",
	), t)
	assert(strings.HasSuffix(
		getFunctionName(tm.NewEditedChannelPostRoute(tm.Any(), func(u *tm.Update) {}).filter),
		"And.func1",
	), t)

	update = &tm.Update{
		Update: tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "Here is a fraction: 3/5. Parse this!",
			},
		},
		Context: map[string]interface{}{},
	}
	tm.NewRegexRoute("([0-9]+)/([1-9][0-9]*)", nil).Handle(nil, update)
	assert(reflect.DeepEqual(update.Context["matches"], []string{"3/5", "3", "5"}), t)
}
