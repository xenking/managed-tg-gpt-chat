package tgrouter

import (
	"context"
)

type RouteGroup interface {
	Route
	Use(middlewares ...Middleware)
}

func NewGroup(filter FilterMatcher, routes ...Route) RouteGroup {
	return &routeGroup{
		filter: filter,
		routes: routes,
	}
}

func NewGroupAny(routes ...Route) RouteGroup {
	return NewGroup(Any(), routes...)
}

type routeGroup struct {
	filter      FilterMatcher
	routes      []Route
	middlewares []Middleware
}

func (g *routeGroup) Handle(ctx context.Context, u *Update) error {
	route := g.matchRoute(u)
	if route == nil {
		return ErrRouteNotFound
	}
	wh := g.wrap(route)
	return wh.Handle(ctx, u)
}

func (g *routeGroup) Match(u *Update) bool {
	return g.filter.Match(u)
}

func (g *routeGroup) Use(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *routeGroup) wrap(handler Handler) Handler {
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		handler = g.middlewares[i](handler)
	}
	return handler
}

func (g *routeGroup) matchRoute(u *Update) Handler {
	// TODO: use more efficient search algorithm like Aho-Corasick
	for _, route := range g.routes {
		if route.Match(u) {
			return route
		}
	}
	return nil
}
