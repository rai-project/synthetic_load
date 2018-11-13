package synthetic_load

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/seehuhn/mt19937"
)

type TraceEntry struct {
	TimeStamp      time.Duration
	GeneratorIndex int
}

type Trace []TraceEntry

// Generate a trace from a query library based on a seed with a given minimum
// number of queries, miniumum duration, and qps.
func NewTrace(opts ...Option) Trace {
	options := NewOptions(opts...)

	// Using the std::mt19937 pseudo-random number generator ensures a modicum of
	// cross platform reproducibility for trace generation.
	mt := mt19937.New()
	mt.Seed(options.seed)
	rng := rand.New(mt)

	timeStamp := time.Duration(0)
	tr := []TraceEntry{}

	for timeStamp < options.minDuration || len(tr) < options.minQueries {
		// Poisson arrival process corresponds to exponentially distributed
		// interarrival times.
		timeStamp += time.Duration(rng.ExpFloat64() * float64(time.Millisecond))
		tr = append(tr,
			TraceEntry{
				TimeStamp:      timeStamp,
				GeneratorIndex: rand.Int(),
			},
		)
	}

	return Trace(tr)
}

func (trace Trace) QPS() float64 {
	traceLength := len(trace)
	last := trace[traceLength-1]
	first := trace[0]
	duration := last.TimeStamp - first.TimeStamp
	return float64(duration.Nanoseconds()) / float64(traceLength)
}

// Replay a trace using a user provided work enqueueing function. Returns the
// 99-percentile latency.
func (trace Trace) Replay(opts ...Option) time.Duration {
	options := NewOptions(opts...)

	if len(trace) == 0 {
		return time.Duration(0)
	}

	latencies := make([]time.Duration, len(trace))
	start := time.Now()

	for ii := range trace {
		ii := ii
		tr := trace[ii]
		go func() {
			queryStartTime := start.Add(tr.TimeStamp)
			time.Sleep(tr.TimeStamp)
			input, err := options.inputGenerator(tr.GeneratorIndex)
			if err != nil {
				log.WithError(err).Panic("unable to generate input")
			}
			options.runner.Run(
				input,
				func() {
					latencies[ii] = time.Since(queryStartTime)
				},
			)
		}()
	}

	sort.Slice(latencies, func(ii, jj int) bool {
		return latencies[ii] < latencies[jj]
	})

	idx := int(math.Ceil(options.latencyBoundPercentile * float64(len(latencies)-1)))
	return latencies[idx]
}

// Returns the maximum throughput (QPS) subject to a 99-percentile latency
// bound.
func FindMaxQPS(opts ...Option) float64 {

	options := NewOptions(opts...)

	qpsLowerBound := 0.0
	qpsUpperBound := math.MaxFloat64

	iters := int64(0)
	relativeQpsTolerance := 0.01

	for (qpsUpperBound-qpsLowerBound)/qpsLowerBound > relativeQpsTolerance && iters < options.maxQpsSearchIterations {
		iters++
		targetQps := 0.0
		if qpsLowerBound == 0 && qpsUpperBound == math.MaxFloat64 {
			targetQps = 512
		} else if qpsUpperBound == math.MaxFloat64 {
			targetQps = 2 * qpsLowerBound
		} else {
			targetQps = (qpsLowerBound + qpsUpperBound) / 2
		}

		log.WithField("targetQps", targetQps).Debug("creating a new trace")

		options.seed += 1
		trace := NewTrace(append(opts, Seed(options.seed), QPS(targetQps))...)
		traceQps := trace.QPS()
		if qpsLowerBound < traceQps && traceQps < qpsUpperBound {
			log.Debug("replaying trace")
			measuredLatency := trace.Replay(opts...)

			log.WithField("qps", traceQps).
				WithField("latency_bound_percentile", 100*options.latencyBoundPercentile).
				WithField("% latency", measuredLatency).
				Info("replayed trace")
			if measuredLatency > options.latencyBound {
				qpsUpperBound = math.Min(qpsUpperBound, traceQps)
			} else {
				qpsLowerBound = math.Max(traceQps, qpsLowerBound)
			}
		}

		log.WithField("qpsUpperBound", qpsUpperBound).
			WithField("qpsLowerBound", qpsLowerBound).
			Trace("generated new trace")
	}
	return math.Min(qpsUpperBound, qpsLowerBound)
}
