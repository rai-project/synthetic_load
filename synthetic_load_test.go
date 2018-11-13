package synthetic_load

import (
	"testing"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
)

func TestSleepingRunner(t *testing.T) {
	qps := FindMaxQPS(MaxQPSSearchIterations(100))
	assert.NotEmpty(t, qps)

	pp.Println(qps)
}
