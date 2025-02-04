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
	ErrorCount          int64
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

func (s *Stats) IncrementErrors() {
	atomic.AddInt64(&s.ErrorCount, 1)
}

type Progress struct {
	totalLines     int64
	processedLines int64
	lastDisplay    time.Time
	startTime      time.Time
	stats          *Stats
	messageBuffer  []string
}

func NewProgress(total int, stats *Stats) *Progress {
	return &Progress{
		totalLines:    int64(total),
		lastDisplay:   time.Now(),
		startTime:     time.Now(),
		stats:         stats,
		messageBuffer: make([]string, 0),
	}
}

func (p *Progress) Increment() {
	atomic.AddInt64(&p.processedLines, 1)
}

func (p *Progress) Processed() int64 {
	return atomic.LoadInt64(&p.processedLines)
}

// BufferMessage stores messages to be displayed periodically.
func (p *Progress) BufferMessage(msg string) {
	p.messageBuffer = append(p.messageBuffer, msg)
}

// FlushMessages displays all buffered messages.
func (p *Progress) FlushMessages() {
	if len(p.messageBuffer) == 0 {
		return
	}
	fmt.Println("\n--- Messages ---")
	for _, msg := range p.messageBuffer {
		fmt.Println(msg)
	}
	p.messageBuffer = p.messageBuffer[:0]
}

// Display outputs multi-line progress information.
func (p *Progress) Display() {
	processed := p.Processed()
	total := atomic.LoadInt64(&p.totalLines)
	if total == 0 {
		return
	}

	ratio := float64(processed) / float64(total)
	percentage := ratio * 100

	// Generate progress bar
	barWidth := 50
	fillCount := int(ratio * float64(barWidth))
	bar := strings.Repeat("=", fillCount) + strings.Repeat("-", barWidth-fillCount)

	// Calculate rate and ETA
	elapsed := time.Since(p.startTime)
	linesPerSec := float64(processed) / elapsed.Seconds()
	remainingLines := total - processed
	eta := time.Duration(float64(remainingLines)/linesPerSec) * time.Second

	// Core progress display (multi-line)
	fmt.Printf("\033[H\033[J") // Clear terminal screen (simple ANSI escape code)
	fmt.Printf(
		"%s\n[%s] %.2f%% (%d/%d lines)\nRate: %.2f lines/sec | ETA: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		bar, percentage, processed, total, linesPerSec, eta.Truncate(time.Second),
	)

	// Detailed stats
	if p.stats != nil {
		eventLines := atomic.LoadInt64(&p.stats.EventLinesProcessed)
		skipped := atomic.LoadInt64(&p.stats.SkippedLines)
		errors := atomic.LoadInt64(&p.stats.ErrorCount)
		alreadyInDB := atomic.LoadInt64(&p.stats.AlreadyInDB)
		newIPs := atomic.LoadInt64(&p.stats.NewIPsDiscovered)

		fmt.Printf(
			"Events: %d | Skipped: %d | Errors: %d | DB (Existing: %d) | New IPs: %d\n",
			eventLines, skipped, errors, alreadyInDB, newIPs,
		)
	}
}

func (p *Progress) DisplayIfNeeded(minInterval time.Duration) {
	now := time.Now()
	if now.Sub(p.lastDisplay) >= minInterval {
		p.FlushMessages()
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
			p.FlushMessages()
			p.Display()
		case <-done:
			p.FlushMessages()
			p.Display()
			fmt.Println("--- Final Summary ---")
			return
		}
	}
}
