package worker

import "time"

type Backoff struct{ Base, Max time.Duration }

func (b Backoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return b.Base
	}
	d := b.Base
	for i := 1; i < attempt; i++ {
		if d >= b.Max/2 {
			return b.Max
		}
		d *= 2
	}
	if d > b.Max {
		return b.Max
	}
	return d
}
func Retryable(status string) bool                                { return status == "queued" || status == "retry" }
func Terminal(status string) bool                                 { return status == "done" || status == "failed" }
func NextAttempt(now time.Time, attempt int, b Backoff) time.Time { return now.Add(b.Delay(attempt)) }
