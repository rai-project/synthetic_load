package synthetic_load

import (
	"testing"
	"time"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
)

func TestSleepingRunner(t *testing.T) {
	qps := FindMaxQPS(
		LatencyBound(100*time.Millisecond),
		LatencyBoundPercentile(0.99),
		MinDuration(1*time.Second),
		MinQueries(1024),
		MaxQPSSearchIterations(10),
	)
	assert.NotEmpty(t, qps)

	pp.Println(qps)
}
