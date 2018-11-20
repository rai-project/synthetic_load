package synthetic_load

import (
	"time"
)

// This example enqueue function just sleeps for 20 ms.
type SleepingRunner struct {
}

func (s SleepingRunner) Run(input []byte, onFinish func()) error {
	time.Sleep(20 * time.Millisecond)
	onFinish()
	return nil
}
