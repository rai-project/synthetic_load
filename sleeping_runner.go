package synthetic_load

import (
	"time"
)

// This example enqueue function just sleeps for 20 ms.
type SleepingRunner struct {
}

func (s SleepingRunner) Run(tr TraceEntry, batchSize int, input []byte, onFinish func()) error {
	time.Sleep(10 * time.Millisecond)
	onFinish()
	return nil
}
