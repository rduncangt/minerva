package progress

import (
	"fmt"
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

func (p *Progress) BufferMessage(msg string) {
	p.messageBuffer = append(p.messageBuffer, msg)
}

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

// Display outputs the current progress status, omitting the progress bar.
func (p *Progress) Display() {
	processed := p.Processed()
	total := atomic.LoadInt64(&p.totalLines)
	if total == 0 {
		return
	}

	percentage := (float64(processed) / float64(total)) * 100
	elapsed := time.Since(p.startTime)
	linesPerSec := float64(processed) / elapsed.Seconds()
	remainingLines := total - processed
	eta := time.Duration(float64(remainingLines)/linesPerSec) * time.Second

	// Core progress display
	fmt.Printf("\033[H\033[J") // Clear terminal screen
	fmt.Printf(
		"%s | %.2f%% Complete (%d/%d lines)\nRate: %.2f lines/sec | ETA: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		percentage, processed, total, linesPerSec, eta.Truncate(time.Second),
	)

	// Display detailed stats
	if p.stats != nil {
		errors := atomic.LoadInt64(&p.stats.ErrorCount)
		eventLines := atomic.LoadInt64(&p.stats.EventLinesProcessed)
		alreadyInDB := atomic.LoadInt64(&p.stats.AlreadyInDB)
		newIPs := atomic.LoadInt64(&p.stats.NewIPsDiscovered)
		skipped := atomic.LoadInt64(&p.stats.SkippedLines)

		fmt.Printf(
			"Errors: %d | Events: %d | DB (Existing: %d) | New IPs: %d | Skipped: %d\n",
			errors, eventLines, alreadyInDB, newIPs, skipped,
		)
	}
}

// FormatDuration formats the elapsed time for display.
func FormatDuration(d time.Duration) string {
	if d >= time.Minute {
		return fmt.Sprintf("%dm %.2fs", int(d.Minutes()), d.Seconds()-float64(int(d.Minutes())*60))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
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
			p.finalSummaryDisplay()
			return
		}
	}
}

// finalSummaryDisplay prints the final summary and JSON output.
func (p *Progress) finalSummaryDisplay() {
	elapsed := time.Since(p.startTime)
	fmt.Printf("\n%s\n--- Final Summary ---\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Execution time: %s\n", FormatDuration(elapsed))

	if p.stats != nil {
		errors := atomic.LoadInt64(&p.stats.ErrorCount)
		eventLines := atomic.LoadInt64(&p.stats.EventLinesProcessed)
		alreadyInDB := atomic.LoadInt64(&p.stats.AlreadyInDB)
		newIPs := atomic.LoadInt64(&p.stats.NewIPsDiscovered)
		skipped := atomic.LoadInt64(&p.stats.SkippedLines)

		fmt.Printf(
			"Errors: %d | Events: %d | DB (Existing: %d) | New IPs: %d | Skipped: %d\n",
			errors, eventLines, alreadyInDB, newIPs, skipped,
		)

		// JSON output embedded in the final summary
		fmt.Printf("\nJSON Output:\n")
		fmt.Printf(
			"{\"total_lines\":%d,\"suspicious_events\":%d,\"records_inserted\":%d,\"new_ips\":%d}\n",
			atomic.LoadInt64(&p.totalLines),
			eventLines,
			alreadyInDB,
			newIPs,
		)
	}
}
