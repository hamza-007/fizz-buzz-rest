package rest

import (
	"context"

	"fizz-buzz-rest/internal/fizzbuzz"
)

// TODO: the seam for a shared recorder (Redis, Postgres), with no handler change.
type StatsRecorder interface {
	Add(fizzbuzz.Request)
	Top() (fizzbuzz.Request, uint64, bool)
}

type Options struct {
	MaxLimit int
	Stats    StatsRecorder
	Health   func(context.Context) error
}

type API struct {
	maxLimit int
	stats    StatsRecorder
	health   func(context.Context) error
}

func New(opts Options) *API {
	if opts.Stats == nil {
		panic("rest: Options.Stats is required")
	}

	health := opts.Health
	if health == nil {
		health = func(context.Context) error { return nil }
	}

	return &API{
		maxLimit: opts.MaxLimit,
		stats:    opts.Stats,
		health:   health,
	}
}
