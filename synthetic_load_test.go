package synthetic_load

import (
	"testing"
	"time"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
)

func TestSleepingRunner(t *testing.T) {
	qps := FindMaxQPS(
		QPS(100),
		LatencyBound(400*time.Millisecond),
		LatencyBoundPercentile(0.99),
		MinDuration(1*time.Second),
		MinQueries(64),
		BatchSize(32),
		MaxQPSSearchIterations(10),
	)
	assert.NotEmpty(t, qps)

	pp.Println(qps)
}

func init() {
	pp.WithLineInfo = true
}
