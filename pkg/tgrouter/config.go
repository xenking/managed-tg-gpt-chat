package tgrouter

type Config struct {
	RecoverHandler  RecoverHandlerFunc
	ErrorHandler    ErrorHandlerFunc
	NotFoundHandler Handler
	GlobalFilter    FilterFunc
}

type Option interface {
	Apply(cfg *Config)
}

type option func(cfg *Config)

func (o option) Apply(cfg *Config) {
	o(cfg)
}

func WithRecoverHandler(r RecoverHandlerFunc) Option {
	return option(func(cfg *Config) {
		cfg.RecoverHandler = r
	})
}

func WithErrorHandler(h ErrorHandlerFunc) Option {
	return option(func(cfg *Config) {
		cfg.ErrorHandler = h
	})
}

func WithNotFoundHandler(h Handler) Option {
	return option(func(cfg *Config) {
		cfg.NotFoundHandler = h
	})
}

func WithGlobalFilter(f FilterFunc) Option {
	return option(func(cfg *Config) {
		cfg.GlobalFilter = f
	})
}
