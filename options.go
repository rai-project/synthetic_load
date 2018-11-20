package synthetic_load

import (
	"context"
	"math"
	"time"
)

type Options struct {
	ctx                    context.Context
	seed                   int64
	minQueries             int
	minDuration            time.Duration
	latencyBound           time.Duration
	latencyBoundPercentile float64
	runner                 Runner
	qps                    float64
	maxQpsSearchIterations int64
}

type Option func(*Options)

func SetContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

// The input runner (what's called enqueue function in sylt)
func SetRunner(runner Runner) Option {
	return func(o *Options) {
		o.runner = runner
	}
}

// The pseudo-random number generator's seed.
func SetSeed(seed int64) Option {
	return func(o *Options) {
		o.seed = seed
	}
}

// The minimum number of queries.
func SetMinQueries(m int) Option {
	return func(o *Options) {
		o.minQueries = m
	}
}

// The minimum duration of the trace.
func SetMinDuration(d time.Duration) Option {
	return func(o *Options) {
		o.minDuration = d
	}
}

func SetQPS(qps float64) Option {
	return func(o *Options) {
		o.qps = qps
	}
}

// The target latency bound.
func SetLatencyBound(latencyBound time.Duration) Option {
	return func(o *Options) {
		o.latencyBound = latencyBound
	}
}

// The minimum percent of queries meeting the latency bound.
func SetLatencyBoundPercentile(latencyBoundPercentile float64) Option {
	return func(o *Options) {
		o.latencyBoundPercentile = latencyBoundPercentile
	}
}

func SetMaxQPSSearchIterations(maxQpsSearchIterations int64) Option {
	return func(o *Options) {
		o.maxQpsSearchIterations = maxQpsSearchIterations
	}
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		ctx:                    context.Background(),
		seed:                   0, //time.Now().UnixNano(),
		latencyBound:           100 * time.Millisecond,
		latencyBoundPercentile: 0.99,
		minDuration:            1 * time.Second,
		minQueries:             1024,
		runner:                 SleepingRunner{},
		maxQpsSearchIterations: math.MaxInt64,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}
