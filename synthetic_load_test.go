package synthetic_load

import (
	"testing"
	"time"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
)

func TestSleepingRunner(t *testing.T) {
	qps := FindMaxQPS(
		SetLatencyBound(100*time.Millisecond),
		SetLatencyBoundPercentile(0.99),
		SetMinDuration(1*time.Second),
		SetMinQueries(1024),
		SetMaxQPSSearchIterations(10),
	)
	assert.NotEmpty(t, qps)

	pp.Println(qps)
}
