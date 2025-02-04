package progress

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type Stats struct {
	EventLinesProcessed int64
	SkippedLines        int64
	AlreadyInDB         int64
	NewIPsDiscovered    int64
}

// Increment methods for each stat.
func (s *Stats) IncrementEventLines() {
	atomic.AddInt64(&s.EventLinesProcessed, 1)
}

func (s *Stats) IncrementSkippedLines() {
	atomic.AddInt64(&s.SkippedLines, 1)
}

func (s *Stats) IncrementAlreadyInDB() {
	atomic.AddInt64(&s.AlreadyInDB, 1)
}

func (s *Stats) IncrementNewIPs() {
	atomic.AddInt64(&s.NewIPsDiscovered, 1)
}

type Progress struct {
	totalLines     int64
	processedLines int64
	lastDisplay    time.Time
	startTime      time.Time
	stats          *Stats
}

func NewProgress(total int, stats *Stats) *Progress {
	return &Progress{
		totalLines:  int64(total),
		lastDisplay: time.Now(),
		startTime:   time.Now(),
		stats:       stats,
	}
}

func (p *Progress) Increment() {
	atomic.AddInt64(&p.processedLines, 1)
}

func (p *Progress) Processed() int64 {
	return atomic.LoadInt64(&p.processedLines)
}

func (p *Progress) Display() {
	processed := p.Processed()
	total := atomic.LoadInt64(&p.totalLines)
	if total == 0 {
		return
	}

	ratio := float64(processed) / float64(total)
	percentage := ratio * 100

	// Generate an ASCII progress bar
	barWidth := 50
	fillCount := int(ratio * float64(barWidth))
	bar := strings.Repeat("=", fillCount) + strings.Repeat("-", barWidth-fillCount)

	// Calculate processing rate and ETA
	elapsed := time.Since(p.startTime)
	linesPerSec := float64(processed) / elapsed.Seconds()
	remainingLines := total - processed
	eta := time.Duration(float64(remainingLines)/linesPerSec) * time.Second

	// Build the core progress line
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf(
		"\r%s [%s] %.2f%% (%d/%d lines) | Rate: %.2f lines/sec | ETA: %s",
		timestamp, bar, percentage, processed, total, linesPerSec, eta.Truncate(time.Second),
	)

	// Append detailed stats
	if p.stats != nil {
		eventLines := atomic.LoadInt64(&p.stats.EventLinesProcessed)
		skipped := atomic.LoadInt64(&p.stats.SkippedLines)
		alreadyInDB := atomic.LoadInt64(&p.stats.AlreadyInDB)
		newIPs := atomic.LoadInt64(&p.stats.NewIPsDiscovered)

		line += fmt.Sprintf(
			" | Events: %d | Skipped: %d | DB (Existing: %d) | New IPs: %d",
			eventLines, skipped, alreadyInDB, newIPs,
		)
	}

	// Print the line in place
	fmt.Print(line)
}

func (p *Progress) DisplayIfNeeded(minInterval time.Duration) {
	now := time.Now()
	if now.Sub(p.lastDisplay) >= minInterval {
		p.Display()
		p.lastDisplay = now
	}
}

func (p *Progress) StartPeriodicDisplay(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.Display()
		case <-done:
			// One final display and newline for cleanliness
			p.Display()
			fmt.Println()
			return
		}
	}
}
