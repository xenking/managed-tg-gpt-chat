package tgrouter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
)

var ErrRouteNotFound = errors.New("route not found")

type FilterMatcher interface {
	Match(u *Update) bool
}

// Handler is either routeHandler or Router.
type Handler interface {
	Handle(ctx context.Context, u *Update) error
}

// Route is a combination of FilterMatcher and Handler.
type Route interface {
	FilterMatcher
	Handler
}

// RecoverHandlerFunc handles panics which happen during Dispatch.
type RecoverHandlerFunc = func(u *Update, err error)

// ErrorHandlerFunc handles errors during routeHandler.Handle
type ErrorHandlerFunc func(ctx context.Context, u *Update, err error)

// HandlerFunc handles tg update
type HandlerFunc func(ctx context.Context, u *Update) error

func (fn HandlerFunc) Handle(ctx context.Context, u *Update) error {
	return fn(ctx, u)
}

type Middleware func(handler Handler) Handler

type Router struct {
	api         echotron.API
	botSelf     *echotron.User
	botSelfOnce sync.Once
	routes      []Route // Contains instances of Router & Handler
	cfg         Config
}

// NewRouter creates new multiplexer.
func NewRouter(api echotron.API, opts ...Option) *Router {
	r := &Router{
		api: api,
	}
	for _, opt := range opts {
		opt.Apply(&r.cfg)
	}
	return r
}

// Mount adds one or more handlers to router.
func (r *Router) Mount(routes ...Route) *Router {
	r.routes = append(r.routes, routes...)
	return r
}

func (r *Router) tryRecover(u *Update) {
	if p := recover(); p != nil {
		err, ok := p.(error)
		if !ok {
			err = fmt.Errorf("%v", p)
		}
		if r.cfg.RecoverHandler != nil {
			r.cfg.RecoverHandler(u, err)
		} else {
			panic(err)
		}
	}
}

// Handle runs router with provided update.
func (r *Router) Handle(ctx context.Context, u *Update) error {
	defer r.tryRecover(u)

	if r.cfg.GlobalFilter != nil && !r.Match(u) {
		return nil
	}

	route := r.matchRoute(u)
	if route == nil {
		if r.cfg.NotFoundHandler != nil {
			return r.cfg.NotFoundHandler.Handle(ctx, u)
		}
		return ErrRouteNotFound
	}

	err := route.Handle(ctx, u)
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrRouteNotFound) && r.cfg.NotFoundHandler != nil {
		return r.cfg.NotFoundHandler.Handle(ctx, u)
	}
	if r.cfg.ErrorHandler != nil {
		r.cfg.ErrorHandler(ctx, u, err)
	}
	return nil
}

func (r *Router) matchRoute(u *Update) Handler {
	// TODO: use more efficient search algorithm like Aho-Corasick
	for _, route := range r.routes {
		if route.Match(u) {
			return route
		}
	}
	return r.cfg.NotFoundHandler
}

func (r *Router) Match(u *Update) bool {
	return r.cfg.GlobalFilter != nil && r.cfg.GlobalFilter(u)
}

// HandleUpdate for embedding into echotron dispatcher
func (r *Router) HandleUpdate(ctx context.Context, u *echotron.Update) {
	err := r.getBotSelf(ctx)
	upd := NewUpdate(u, r.api, r.botSelf)
	if err != nil {
		r.cfg.ErrorHandler(ctx, upd, err)
	}
	if hErr := r.Handle(ctx, upd); hErr != nil {
		r.cfg.ErrorHandler(ctx, upd, hErr)
	}
}

func (r *Router) getBotSelf(ctx context.Context) (err error) {
	r.botSelfOnce.Do(func() {
		var resp echotron.APIResponseUser
		resp, err = r.api.GetMe(ctx)
		r.botSelf = resp.Result
	})
	return err
}
