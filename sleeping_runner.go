package synthetic_load

import (
	"time"
)

// This example enqueue function just sleeps for 100 us.
type SleepingRunner struct {
}

func (s SleepingRunner) Run(input []byte, onFinish func()) error {
	time.Sleep(100 * time.Microsecond)
	onFinish()
	return nil
}
