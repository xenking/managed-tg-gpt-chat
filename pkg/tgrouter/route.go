package tgrouter

import (
	"context"
	"regexp"
	"strings"
)

// routeHandler defines a function that will handle updates that pass the filtering.
type routeHandler struct {
	filter      FilterMatcher
	handler     Handler
	middlewares []Middleware
}

func (h *routeHandler) Handle(ctx context.Context, u *Update) error {
	wh := h.wrap(h.handler)
	return wh.Handle(ctx, u)
}

func (h *routeHandler) Match(u *Update) bool {
	return h.filter.Match(u)
}

func (h *routeHandler) Use(middlewares ...Middleware) {
	h.middlewares = append(h.middlewares, middlewares...)
}

func (h *routeHandler) wrap(handler Handler) Handler {
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}
	return handler
}

// NewRoute creates a new generic routeHandler.
func NewRoute(filter FilterMatcher, handler Handler, middlewares ...Middleware) Route {
	if filter == nil {
		filter = Any()
	}
	return &routeHandler{
		filter:      filter,
		handler:     handler,
		middlewares: middlewares,
	}
}

func NewAnyRoute(handlers Handler) Route {
	return NewRoute(Any(), handlers)
}

// NewMessageRoute creates a routeHandler for updates that contain message.
func NewMessageRoute(filter FilterMatcher, handler Handler) Route {
	newFilter := IsMessage()
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	return NewRoute(newFilter, handler)
}

// NewRegexRoute creates a routeHandler for updates that contain message which matches the pattern as regexp.
func NewRegexRoute(pattern string, filter FilterMatcher, handler Handler) Route {
	exp := regexp.MustCompile(pattern)
	newFilter := And(IsMessage(), FilterFunc(func(u *Update) bool {
		return exp.Match([]byte(u.Message.Text))
	}))
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	//handles = append([]HandlerFunc{
	//	func(u *Update) {
	//		u.Context["exp"] = exp
	//		u.Context["matches"] = exp.FindStringSubmatch(u.Message.Text)
	//	},
	//}, handles...)
	return NewRoute(newFilter, handler)
}

// NewCommandRoute is an extension for NewMessageRoute that creates a routeHandler for updates that contain message with command.
// It also populates u.Context["args"] with a slice of strings.
//
// For example, when invoked as `/somecmd foo bar 1337`, u.Context["args"] will be set to []string{"foo", "bar", "1337"}
//
// command can be a string (like "start" or "somecmd") or a space-delimited list of commands to accept (like "start somecmd othercmd")
func NewCommandRoute(command string, filter FilterMatcher, handler Handler) Route {
	var commandFilters []FilterMatcher
	for _, variant := range strings.Split(command, " ") {
		commandFilters = append(commandFilters, IsCommandMessage(strings.TrimPrefix(variant, "/")))
	}
	newFilter := Or(commandFilters...)
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	//handles = append([]HandlerFunc{
	//	func(u *Update) {
	//		u.Context["args"] = strings.Split(u.Message.Text, " ")[1:]
	//	},
	//}, handles...)
	return NewMessageRoute(
		newFilter,
		handler,
	)
}

// NewInlineQueryRoute creates a routeHandler for updates that contain inline query which matches the pattern as regexp.
func NewInlineQueryRoute(pattern string, filter FilterMatcher, handler Handler) Route {
	exp := regexp.MustCompile(pattern)
	newFilter := And(IsInlineQuery(), FilterFunc(func(u *Update) bool {
		return exp.Match([]byte(u.InlineQuery.Query))
	}))
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	//handles = append([]HandlerFunc{
	//	func(u *Update) {
	//		u.Context["exp"] = exp
	//		u.Context["matches"] = exp.FindStringSubmatch(u.InlineQuery.Query)
	//	},
	//}, handles...)
	return NewRoute(newFilter, handler)
}

// NewCallbackQueryHandler creates a routeHandler for updates that contain callback query which matches the pattern as regexp.
func NewCallbackQueryHandler(pattern string, filter FilterMatcher, handler Handler) Route {
	exp := regexp.MustCompile(pattern)
	newFilter := And(IsCallbackQuery(), FilterFunc(func(u *Update) bool {
		return exp.Match([]byte(u.CallbackQuery.Data))
	}))
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	//handles = append([]HandlerFunc{
	//	func(u *Update) {
	//		u.Context["exp"] = exp
	//		u.Context["matches"] = exp.FindStringSubmatch(u.CallbackQuery.Data)
	//	},
	//}, handles...)
	return NewRoute(newFilter, handler)
}

// NewEditedMessageRoute creates a routeHandler for updates that contain edited message.
func NewEditedMessageRoute(filter FilterMatcher, handler Handler) Route {
	newFilter := IsEditedMessage()
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	return NewRoute(newFilter, handler)
}

// NewChannelPostRoute creates a routeHandler for updates that contain channel post.
func NewChannelPostRoute(filter FilterMatcher, handler Handler) Route {
	newFilter := IsChannelPost()
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	return NewRoute(newFilter, handler)
}

// NewEditedChannelPostRoute creates a routeHandler for updates that contain edited channel post.
func NewEditedChannelPostRoute(filter FilterMatcher, handler Handler) Route {
	newFilter := IsEditedChannelPost()
	if filter != nil {
		newFilter = And(newFilter, filter)
	}
	return NewRoute(newFilter, handler)
}

// StateMap is an alias to map of strings to routeHandler slices.
type StateMap map[string]Route

// NewConversationRoute creates a conversation routeHandler.
//
// "conversationID" distinguishes this conversation from the others. The main goal of this identifier is to allow persistence to keep track of different conversation states independently without mixing them together.
//
// "persistence" defines where to store conversation state & intermediate inputs from the user. Without persistence, a conversation would not be able to "remember" what "step" the user is at.
//
// "states" define what handlers to use in which state. States are usually strings like "upload_photo", "send_confirmation", "wait_for_text" and describe the "step" the user is currently at.
// Empty string (`""`) should be used as an initial/final state (i. e. if the conversation has not started yet or has already finished.)
// For each state you must provide a slice with at least one routeHandler. If none of the handlers can handle the update, the default handlers are attempted (see below).
// In order to switch to a different state your routeHandler must call `u.PersistenceContext.SetState("STATE_NAME") ` replacing STATE_NAME with the name of the state you want to switch into.
// Conversation data can be accessed with `u.PersistenceContext.GetData()` and updated with `u.PersistenceContext.SetData(newData)`.
//
// "defaults" are "appended" to every state except default state (`""`). They are useful to handle commands such as "/cancel" or to display some default message.
func NewConversationRoute(
	conversationID string,
	persistence ConversationPersistence,
	states StateMap,
	defaultHandler Handler,
) Route {
	return NewRoute(
		FilterFunc(func(u *Update) bool {
			user, chat := u.EffectiveUser(), u.EffectiveChat()
			if user == nil || chat == nil {
				return false
			}
			pk := PersistenceKey{conversationID, user.ID, chat.ID}
			state := persistence.GetState(pk)
			route := states[state]
			u.PersistenceContext = &PersistenceContext{
				Persistence: persistence,
				PK:          pk,
			}
			defer func() { u.PersistenceContext = nil }()
			return route.Match(u)
		}),
		HandlerFunc(func(ctx context.Context, u *Update) (err error) {
			user, chat := u.EffectiveUser(), u.EffectiveChat()
			pk := PersistenceKey{conversationID, user.ID, chat.ID}
			state := persistence.GetState(pk)
			route := states[state]
			if route == nil {
				route = NewAnyRoute(defaultHandler)
			}
			if u.PersistenceContext == nil {
				u.PersistenceContext = &PersistenceContext{
					Persistence: persistence,
					PK:          pk,
				}
				defer func() { u.PersistenceContext = nil }()
			}
			defer func() {
				if u.PersistenceContext.NewState != nil {
					// TODO: Add docs for :enter hook
					if nextRoute, ok := states[*u.PersistenceContext.NewState+":enter"]; ok {
						err = nextRoute.Handle(ctx, u)
					}
				}
			}()
			err = route.Handle(ctx, u)
			return err
		}),
	)
}
