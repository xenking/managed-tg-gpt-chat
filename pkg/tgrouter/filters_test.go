package tgrouter_test

import (
	"fmt"
	"testing"

	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
	tm "github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

func TestIsCommandMessage(t *testing.T) {
	Check := func(text string, isCommand bool) {
		u := &tm.Update{}
		u.Update.Message = &echotron.Message{}
		u.Update.Message.Text = text
		u.API = &echotron.API{}
		u.API.Self.UserName = "testbot"
		actual := tm.IsCommandMessage("foo")(u)
		if actual != isCommand {
			t.Errorf("Testing %s: IsCommandMessage = %v, expected = %v", text, actual, isCommand)
		}
	}

	Check("asd", false)
	Check("/", false)
	Check("/foo", true)
	Check("/foo@", true)
	Check("/foo ", true)
	Check("/foo bar", true)
	Check("/foox", false)
	Check("/fo", false)
	Check("/foo@testbot", true)
	Check("/foo@testbot bar", true)
	Check("/foo@ bar", true)
	Check("/foo@nope", false)
	Check("/foo@nope bar", false)
	Check("/bar", false)
	Check("/bar baz", false)
	Check("/bar@testbot baz", false)
}

func TestIsAnyCommandMessage(t *testing.T) {
	Check := func(text string, isCommand bool) {
		u := &tm.Update{}
		u.Update.Message = &echotron.Message{}
		u.Update.Message.Text = text
		u.API = &echotron.API{}
		u.BotSelf.UserName = "testbot"
		actual := tm.IsAnyCommandMessage()(u)
		if actual != isCommand {
			t.Errorf("Testing %s: IsAnyCommandMessage = %v, expected = %v", text, actual, isCommand)
		}
	}

	Check("asd", false)
	Check("/", false)
	Check("/foo", true)
	Check("/foo@", true)
	Check("/foo ", true)
	Check("/foo bar", true)
	Check("/foo@testbot", true)
	Check("/foo@testbot bar", true)
	Check("/foo@ bar", true)
	Check("/foo@nope", false)
	Check("/foo@nope bar", false)
}

func TestUpdateTypeFilters(t *testing.T) {
	u := &tm.Update{}
	assert(!tm.IsInlineQuery()(u), t)
	u.InlineQuery = &echotron.InlineQuery{}
	assert(tm.IsInlineQuery()(u), t)

	u = &tm.Update{}
	assert(!tm.IsCallbackQuery()(u), t)
	u.CallbackQuery = &echotron.CallbackQuery{}
	assert(tm.IsCallbackQuery()(u), t)

	u = &tm.Update{}
	assert(!tm.IsEditedMessage()(u), t)
	u.EditedMessage = &echotron.Message{}
	assert(tm.IsEditedMessage()(u), t)

	u = &tm.Update{}
	assert(!tm.IsChannelPost()(u), t)
	u.ChannelPost = &echotron.Message{}
	assert(tm.IsChannelPost()(u), t)

	u = &tm.Update{}
	assert(!tm.IsEditedChannelPost()(u), t)
	u.EditedChannelPost = &echotron.Message{}
	assert(tm.IsEditedChannelPost()(u), t)
}

func TestContentFilters(t *testing.T) {
	u := &tm.Update{}
	assert(!tm.HasText()(u), t)
	u.Message = &echotron.Message{Text: "asd"}
	assert(tm.HasText()(u), t)

	u = &tm.Update{}
	assert(!tm.HasPhoto()(u), t)
	u.Message = &echotron.Message{Photo: []echotron.PhotoSize{}}
	assert(tm.HasPhoto()(u), t)

	u = &tm.Update{}
	assert(!tm.HasVoice()(u), t)
	u.Message = &echotron.Message{Voice: &echotron.Voice{}}
	assert(tm.HasVoice()(u), t)

	u = &tm.Update{}
	assert(!tm.HasAudio()(u), t)
	u.Message = &echotron.Message{Audio: &echotron.Audio{}}
	assert(tm.HasAudio()(u), t)

	u = &tm.Update{}
	assert(!tm.HasAnimation()(u), t)
	u.Message = &echotron.Message{Animation: &echotron.Animation{}}
	assert(tm.HasAnimation()(u), t)

	u = &tm.Update{}
	assert(!tm.HasDocument()(u), t)
	u.Message = &echotron.Message{Document: &echotron.Document{}}
	assert(tm.HasDocument()(u), t)

	u = &tm.Update{}
	assert(!tm.HasSticker()(u), t)
	u.Message = &echotron.Message{Sticker: &echotron.Sticker{}}
	assert(tm.HasSticker()(u), t)

	u = &tm.Update{}
	assert(!tm.HasVideo()(u), t)
	u.Message = &echotron.Message{Video: &echotron.Video{}}
	assert(tm.HasVideo()(u), t)

	u = &tm.Update{}
	assert(!tm.HasVideoNote()(u), t)
	u.Message = &echotron.Message{VideoNote: &echotron.VideoNote{}}
	assert(tm.HasVideoNote()(u), t)

	u = &tm.Update{}
	assert(!tm.HasContact()(u), t)
	u.Message = &echotron.Message{Contact: &echotron.Contact{}}
	assert(tm.HasContact()(u), t)

	u = &tm.Update{}
	assert(!tm.HasLocation()(u), t)
	u.Message = &echotron.Message{Location: &echotron.Location{}}
	assert(tm.HasLocation()(u), t)

	u = &tm.Update{}
	assert(!tm.HasVenue()(u), t)
	u.Message = &echotron.Message{Venue: &echotron.Venue{}}
	assert(tm.HasVenue()(u), t)
}

func TestUpdateChatType(t *testing.T) {
	u := &tm.Update{}
	u.Message = &echotron.Message{}

	assert(!tm.IsPrivate()(u), t)
	assert(!tm.IsGroup()(u), t)
	assert(!tm.IsSuperGroup()(u), t)
	assert(!tm.IsGroupOrSuperGroup()(u), t)
	assert(!tm.IsChannel()(u), t)

	u.Message.Chat = &echotron.Chat{}

	assert(!tm.IsPrivate()(u), t)
	assert(!tm.IsGroup()(u), t)
	assert(!tm.IsSuperGroup()(u), t)
	assert(!tm.IsGroupOrSuperGroup()(u), t)
	assert(!tm.IsChannel()(u), t)

	u.Message.Chat.Type = "private"
	assert(tm.IsPrivate()(u), t)
	u.Message.Chat.Type = "group"
	assert(tm.IsGroup()(u), t)
	assert(tm.IsGroupOrSuperGroup()(u), t)
	u.Message.Chat.Type = "supergroup"
	assert(tm.IsSuperGroup()(u), t)
	assert(tm.IsGroupOrSuperGroup()(u), t)
	u.Message.Chat.Type = "channel"
	assert(tm.IsChannel()(u), t)
}

func TestUpdateMembers(t *testing.T) {
	u := &tm.Update{}
	assert(!tm.IsNewChatMembers()(u), t)
	u.Message = &echotron.Message{}
	assert(!tm.IsNewChatMembers()(u), t)
	u.Message.NewChatMembers = []echotron.User{{}}
	assert(tm.IsNewChatMembers()(u), t)

	u = &tm.Update{}
	assert(!tm.IsLeftChatMember()(u), t)
	u.Message = &echotron.Message{}
	assert(!tm.IsLeftChatMember()(u), t)
	u.Message.LeftChatMember = &echotron.User{}
	assert(tm.IsLeftChatMember()(u), t)
}

func TestCombinationFilters(t *testing.T) {
	u := &tm.Update{}
	for _, test := range []struct {
		a bool
		b bool
		r bool
	}{
		{false, false, false},
		{false, true, false},
		{true, false, false},
		{true, true, true},
	} {
		actual := tm.And(func(u *tm.Update) bool { return test.a }, func(u *tm.Update) bool { return test.b })(u)
		assert(actual == test.r, t, fmt.Sprintf("And(%v, %v) should be %v, got %v", test.a, test.b, test.r, actual))
	}
	for _, test := range []struct {
		a bool
		b bool
		r bool
	}{
		{false, false, false},
		{false, true, true},
		{true, false, true},
		{true, true, true},
	} {
		actual := tm.Or(func(u *tm.Update) bool { return test.a }, func(u *tm.Update) bool { return test.b })(u)
		assert(actual == test.r, t, fmt.Sprintf("Or(%v, %v) should be %v, got %v", test.a, test.b, test.r, actual))
	}
	assert(!tm.Not(tm.Any())(u), t)
	assert(tm.Not(tm.Not(tm.Any()))(u), t)
}
