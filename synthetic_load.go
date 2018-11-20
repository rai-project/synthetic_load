package synthetic_load

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/seehuhn/mt19937"
)

type TraceEntry struct {
	Index      int
	InputIndex int
	TimeStamp  time.Duration
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
		// pp.Println(options.minDuration, " === ", time.Duration((rng.ExpFloat64()/options.qps)*float64(time.Second)))
		timeStamp += time.Duration((rng.ExpFloat64() / options.qps) * float64(time.Second))
		tr = append(tr,
			TraceEntry{
				Index:      len(tr),
				InputIndex: rand.Int(),
				TimeStamp:  timeStamp,
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
	return float64(traceLength) / float64(duration.Seconds())
}

// Replay a trace using a user provided work enqueueing function.
func (trace Trace) Replay(opts ...Option) (time.Duration, error) {
	options := NewOptions(opts...)
	batchSize := options.batchSize

	if len(trace) == 0 {
		return time.Duration(0), errors.New("no traces")
	}
	cnt := len(trace) / batchSize
	if cnt == 0 {
		return time.Duration(0), errors.New("no batches")
	}
	batchLatencies := make([]time.Duration, cnt)

	var wg sync.WaitGroup
	wg.Add(cnt)
	start := time.Now()

	for ii := 0; ii < cnt; ii++ {
		ii := ii
		tr := trace[(ii+1)*batchSize-1]
		go func() {
			defer wg.Done()
			queryStartTime := start.Add(tr.TimeStamp)
			time.Sleep(tr.TimeStamp)
			input, err := options.inputGenerator(tr.InputIndex, batchSize)
			if err != nil {
				log.WithError(err).Panic("unable to generate input")
			}
			options.runner.Run(
				tr,
				batchSize,
				input,
				func() {
					batchLatencies[ii] = time.Since(queryStartTime)
					// fmt.Printf("it took %v to run ii = %v\n", batchLatencies[ii], ii)
				},
			)
		}()
	}

	wg.Wait()

	latencies := make([]time.Duration, cnt*batchSize)
	for ii := range trace {
		tail := (ii/batchSize+1)*batchSize - 1
		if tail < len(trace) {
			latencies[ii] = trace[tail].TimeStamp - trace[ii].TimeStamp + batchLatencies[ii/batchSize]
		}
	}

	sort.Slice(latencies, func(ii, jj int) bool {
		return latencies[ii] < latencies[jj]
	})

	idx := int(math.Ceil(options.latencyBoundPercentile * float64(len(latencies)-1)))

	return latencies[idx], nil
}

// Returns the maximum throughput (QPS) subject to a latency bound.
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
			targetQps = options.qps
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
			measuredLatency, err := trace.Replay(opts...)
			if err != nil {
				break
			}
			fmt.Printf("len(trace) = %v, qps = %v, latency_bound_percentile = %v, latency = %v\n",
				len(trace),
				traceQps,
				100*options.latencyBoundPercentile,
				measuredLatency,
			)
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

		// fmt.Printf("qpsLowerBound = %v qpsUpperBound =%v traceQps =%v\n", qpsLowerBound, qpsUpperBound, traceQps)

		log.WithField("qpsUpperBound", qpsUpperBound).
			WithField("qpsLowerBound", qpsLowerBound).
			Debug("generated new trace")
	}

	return math.Min(qpsUpperBound, qpsLowerBound)
}
