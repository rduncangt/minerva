package progress

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Progress struct {
	totalLines     int64
	processedLines int64
	lastDisplay    time.Time
}

// NewProgress initializes a new Progress tracker.
func NewProgress(total int) *Progress {
	return &Progress{
		totalLines:  int64(total),
		lastDisplay: time.Now(),
	}
}

// Increment increments the count of processed lines.
func (p *Progress) Increment() {
	atomic.AddInt64(&p.processedLines, 1)
}

// Processed returns the number of processed lines.
func (p *Progress) Processed() int64 {
	return atomic.LoadInt64(&p.processedLines)
}

// Display shows the current progress with a timestamp.
func (p *Progress) Display() {
	processed := p.Processed()
	total := atomic.LoadInt64(&p.totalLines)
	if total > 0 {
		percentage := (float64(processed) / float64(total)) * 100
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("%s - %.2f%% (%d/%d lines processed)\n", timestamp, percentage, processed, total)
	}
}

// DisplayIfNeeded only shows progress if a minimum interval has passed.
func (p *Progress) DisplayIfNeeded(minInterval time.Duration) {
	now := time.Now()
	if now.Sub(p.lastDisplay) >= minInterval {
		p.Display()
		p.lastDisplay = now
	}
}

// StartPeriodicDisplay starts periodic progress updates.
func (p *Progress) StartPeriodicDisplay(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.Display()
		case <-done:
			return
		}
	}
}
