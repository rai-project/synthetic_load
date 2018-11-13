package synthetic_load

import (
	"testing"
	"time"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
)

func TestSleepingRunner(t *testing.T) {
	qps := FindMaxQPS(
		// MaxQPSSearchIterations(100),
		MinQueries(256),
		MinDuration(100*time.Millisecond),
	)
	assert.NotEmpty(t, qps)

	pp.Println(qps)
}
