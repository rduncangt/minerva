package progress

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// Stats holds additional counters you might want to display during processing.
// Extend it with whatever fields you need, e.g., InsertSuccesses, InsertFailures, etc.
type Stats struct {
	SuspiciousLines  int64
	NonRelevantLines int64
	NewIPs           int64
}

// IncrementSuspicious increments the suspicious log line count by 1.
func (s *Stats) IncrementSuspicious() {
	atomic.AddInt64(&s.SuspiciousLines, 1)
}

// IncrementNonRelevant increments the non-relevant log line count by 1.
func (s *Stats) IncrementNonRelevant() {
	atomic.AddInt64(&s.NonRelevantLines, 1)
}

// IncrementNewIPs increments the "new IPs" counter by 1.
func (s *Stats) IncrementNewIPs() {
	atomic.AddInt64(&s.NewIPs, 1)
}

// Progress tracks total lines, processed lines, and optionally a Stats struct.
type Progress struct {
	totalLines     int64
	processedLines int64
	lastDisplay    time.Time

	stats *Stats // Holds extra counters for display, if desired.
}

// NewProgress initializes a new Progress tracker.
// Pass in a Stats pointer if you want to track extra counters; otherwise use nil.
func NewProgress(total int, stats *Stats) *Progress {
	return &Progress{
		totalLines:  int64(total),
		lastDisplay: time.Now(),
		stats:       stats,
	}
}

// Increment increments the count of processed lines atomically.
func (p *Progress) Increment() {
	atomic.AddInt64(&p.processedLines, 1)
}

// Processed returns the number of processed lines so far.
func (p *Progress) Processed() int64 {
	return atomic.LoadInt64(&p.processedLines)
}

// Display overwrites the same line in the terminal with progress info.
// It includes an ASCII bar plus any extra stats from the Stats struct.
func (p *Progress) Display() {
	processed := p.Processed()
	total := atomic.LoadInt64(&p.totalLines)
	if total == 0 {
		return
	}
	ratio := float64(processed) / float64(total)
	percentage := ratio * 100

	// A simple ASCII bar of width 50
	barWidth := 50
	fillCount := int(ratio * float64(barWidth))
	if fillCount > barWidth {
		fillCount = barWidth
	}
	bar := strings.Repeat("=", fillCount) + strings.Repeat("-", barWidth-fillCount)

	// Build the core progress line
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("\r%s [%s] %.2f%% (%d/%d lines)", timestamp, bar, percentage, processed, total)

	// If we have stats, append them in a compact way
	if p.stats != nil {
		susp := atomic.LoadInt64(&p.stats.SuspiciousLines)
		nonRel := atomic.LoadInt64(&p.stats.NonRelevantLines)
		newIPs := atomic.LoadInt64(&p.stats.NewIPs)
		line += fmt.Sprintf(" | Susp:%d NonRel:%d NewIPs:%d", susp, nonRel, newIPs)
	}

	// Print without a trailing newline so we overwrite in place
	fmt.Print(line)

	// Optionally flush output immediately if needed:
	// _ = os.Stdout.Sync()
}

// DisplayIfNeeded checks the last display time and updates if enough time has passed.
func (p *Progress) DisplayIfNeeded(minInterval time.Duration) {
	now := time.Now()
	if now.Sub(p.lastDisplay) >= minInterval {
		p.Display()
		p.lastDisplay = now
	}
}

// StartPeriodicDisplay calls Display() on a fixed interval until done is closed.
// Once done is received, it does a final display and prints a newline.
func (p *Progress) StartPeriodicDisplay(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.Display()
		case <-done:
			// One final display, then newline so the next shell prompt or logs look clean.
			p.Display()
			fmt.Println()
			return
		}
	}
}
