package tgrouter

import (
	"regexp"
)

// FilterFunc is used to check if this update should be processed by routeHandler.
type FilterFunc func(u *Update) bool

func (fn FilterFunc) Match(u *Update) bool {
	return fn(u)
}

var commandRegex = regexp.MustCompile("^/([0-9a-zA-Z_]+)(@[0-9a-zA-Z_]{3,})?")

// Any tells routeHandler to process all updates.
func Any() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return true
	})
}

// IsMessage filters updates that look like message (text, photo, location etc.)
func IsMessage() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return u.Message != nil
	})
}

// IsInlineQuery filters updates that are callbacks from inline queries.
func IsInlineQuery() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return u.InlineQuery != nil
	})
}

// IsCallbackQuery filters updates that are callbacks from button presses.
func IsCallbackQuery() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return u.CallbackQuery != nil
	})
}

// IsEditedMessage filters updates that are edits to existing messages.
func IsEditedMessage() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return u.EditedMessage != nil
	})
}

// IsChannelPost filters updates that are channel posts.
func IsChannelPost() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return u.ChannelPost != nil
	})
}

// IsEditedChannelPost filters updates that are edits to existing channel posts.
func IsEditedChannelPost() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return u.EditedChannelPost != nil
	})
}

// HasText filters updates that look like text,
// i. e. have some text and do not start with a slash ("/").
func HasText() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Text != "" && message.Text[0] != '/'
	})
}

// IsAnyCommandMessage filters updates that contain a message and look like a command,
// i. e. have some text and start with a slash ("/").
// If command contains bot username, it is also checked.
func IsAnyCommandMessage() FilterMatcher {
	return And(IsMessage(), FilterFunc(func(u *Update) bool {
		matches := commandRegex.FindStringSubmatch(u.Message.Text)
		if len(matches) == 0 {
			return false
		}
		botName := matches[2]
		if botName != "" && botName != "@"+u.BotSelf.UserName {
			return false
		}
		return true
	}))
}

// IsCommandMessage filters updates that contain a specific command.
// For example, IsCommandMessage("start") will handle a "/start" command.
// This will also allow the user to pass arguments, e. g. "/start foo bar".
// Commands in format "/start@bot_name" and "/start@bot_name foo bar" are also supported.
// If command contains bot username, it is also checked.
func IsCommandMessage(cmd string) FilterMatcher {
	return And(IsAnyCommandMessage(), FilterFunc(func(u *Update) bool {
		matches := commandRegex.FindStringSubmatch(u.Message.Text)
		actualCmd := matches[1]
		return actualCmd == cmd
	}))
}

// HasRegex filters updates that match a regular expression.
// For example, HasRegex("^/get_(\d+)$") will handle commands like "/get_42".
func HasRegex(pattern string) FilterMatcher {
	exp := regexp.MustCompile(pattern)
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && exp.MatchString(message.Text)
	})
}

// HasPhoto filters updates that contain a photo.
func HasPhoto() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Photo != nil
	})
}

// HasVoice filters updates that contain a voice message.
func HasVoice() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Voice != nil
	})
}

// HasAudio filters updates that contain an audio.
func HasAudio() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Audio != nil
	})
}

// HasAnimation filters updates that contain an animation.
func HasAnimation() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Animation != nil
	})
}

// HasDocument filters updates that contain a document.
func HasDocument() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Document != nil
	})
}

// HasSticker filters updates that contain a sticker.
func HasSticker() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Sticker != nil
	})
}

// HasVideo filters updates that contain a video.
func HasVideo() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Video != nil
	})
}

// HasVideoNote filters updates that contain a video note.
func HasVideoNote() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.VideoNote != nil
	})
}

// HasContact filters updates that contain a contact.
func HasContact() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Contact != nil
	})
}

// HasLocation filters updates that contain a location.
func HasLocation() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Location != nil
	})
}

// HasVenue filters updates that contain a venue.
func HasVenue() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.Venue != nil
	})
}

// IsPrivate filters updates that are sent in private chats.
func IsPrivate() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if chat := u.EffectiveChat(); chat != nil {
			return chat.IsPrivate()
		}
		return false
	})
}

// IsGroup filters updates that are sent in a group. See also IsGroupOrSuperGroup.
func IsGroup() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if chat := u.EffectiveChat(); chat != nil {
			return chat.IsGroup()
		}
		return false
	})
}

// IsSuperGroup filters updates that are sent in a superbroup. See also IsGroupOrSuperGroup.
func IsSuperGroup() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if chat := u.EffectiveChat(); chat != nil {
			return chat.IsSuperGroup()
		}
		return false
	})
}

// IsGroupOrSuperGroup filters updates that are sent in both groups and supergroups.
func IsGroupOrSuperGroup() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if chat := u.EffectiveChat(); chat != nil {
			return chat.IsGroup() || chat.IsSuperGroup()
		}
		return false
	})
}

// IsChannel filters updates that are sent in channels.
func IsChannel() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if chat := u.EffectiveChat(); chat != nil {
			return chat.IsChannel()
		}
		return false
	})
}

// IsNewChatMembers filters updates that have users in NewChatMembers property.
func IsNewChatMembers() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if message := u.EffectiveMessage(); message != nil {
			return message.NewChatMembers != nil && len(message.NewChatMembers) > 0
		}
		return false
	})
}

// IsLeftChatMember filters updates that have user in LeftChatMember property.
func IsLeftChatMember() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		if message := u.EffectiveMessage(); message != nil {
			return message.LeftChatMember != nil
		}
		return false
	})
}

func IsForwarded() FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.ForwardOrigin != nil
	})
}

func IsForwardOriginType(t string) FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		message := u.EffectiveMessage()
		return message != nil && message.ForwardOrigin != nil && message.ForwardOrigin.Type == t
	})
}

// And filters updates that pass ALL of the provided filters.
func And(filters ...FilterMatcher) FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		for _, filter := range filters {
			if !filter.Match(u) {
				return false
			}
		}
		return true
	})
}

// Or filters updates that pass ANY of the provided filters.
func Or(filters ...FilterMatcher) FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		for _, filter := range filters {
			if filter.Match(u) {
				return true
			}
		}
		return false
	})
}

// Not filters updates that do not pass the provided filter.
func Not(filter FilterMatcher) FilterMatcher {
	return FilterFunc(func(u *Update) bool {
		return !filter.Match(u)
	})
}
