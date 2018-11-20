package synthetic_load

import (
	"context"
	"math"
	"time"
)

type Options struct {
	ctx                    context.Context
	inputGenerator         func(idx int) ([]byte, error)
	seed                   int64
	minQueries             int
	minDuration            time.Duration
	latencyBound           time.Duration
	latencyBoundPercentile float64
	batchSize              int
	runner                 Runner
	qps                    float64
	maxQpsSearchIterations int64
}

type Option func(*Options)

func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

// The input generator (what's called query library in sylt)
func InputGenerator(inputGenerator func(int) ([]byte, error)) Option {
	return func(o *Options) {
		o.inputGenerator = inputGenerator
	}
}

// The input runner (what's called enqueue function in sylt)
func InputRunner(runner Runner) Option {
	return func(o *Options) {
		o.runner = runner
	}
}

// The pseudo-random number generator's seed.
func Seed(seed int64) Option {
	return func(o *Options) {
		o.seed = seed
	}
}

// The minimum number of queries.
func MinQueries(m int) Option {
	return func(o *Options) {
		o.minQueries = m
	}
}

// The minimum duration of the trace.
func MinDuration(d time.Duration) Option {
	return func(o *Options) {
		o.minDuration = d
	}
}

func QPS(qps float64) Option {
	return func(o *Options) {
		o.qps = qps
	}
}

// The target latency bound.
func LatencyBound(latencyBound time.Duration) Option {
	return func(o *Options) {
		o.latencyBound = latencyBound
	}
}

// The minimum percent of queries meeting the latency bound.
func LatencyBoundPercentile(latencyBoundPercentile float64) Option {
	return func(o *Options) {
		o.latencyBoundPercentile = latencyBoundPercentile
	}
}

func BatchSize(batchSize int) Option {
	return func(o *Options) {
		o.batchSize = batchSize
	}
}

func MaxQPSSearchIterations(maxQpsSearchIterations int64) Option {
	return func(o *Options) {
		o.maxQpsSearchIterations = maxQpsSearchIterations
	}
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		ctx: context.Background(),
		inputGenerator: func(idx int) ([]byte, error) {
			if idx%2 == 0 {
				ReadFile("/cat.jpg")
			}
			return ReadFile("/chicken.jpg")
		},
		seed:                   0, //time.Now().UnixNano(),
		latencyBound:           100 * time.Millisecond,
		latencyBoundPercentile: 0.99,
		minDuration:            1 * time.Second,
		minQueries:             1024,
		batchSize:              1,
		runner:                 SleepingRunner{},
		maxQpsSearchIterations: math.MaxInt64,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}
