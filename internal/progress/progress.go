package progress

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Stats holds various counters that we’ll track to inform the user about pipeline progress.
type Stats struct {
	// Basic pipeline counts
	linesRead int64 // number of lines read from stdin

	flagged   int64 // lines that matched some threat/flagging criteria
	benign    int64 // lines that were valid but not flagged
	malformed int64 // lines that can’t be parsed or are incomplete

	inserted int64 // how many were successfully inserted to DB
	errors   int64 // how many errors occurred overall

	// Geo lookup details
	geoQueued    int64 // how many IPs are queued for geo lookup (not processed yet)
	geoCompleted int64 // how many IPs have had geo lookup completed
	geoErrors    int64 // how many IPs failed geo lookup
}

// Atomic incrementers
func (s *Stats) IncrementLinesRead() { atomic.AddInt64(&s.linesRead, 1) }

func (s *Stats) IncrementFlagged() { atomic.AddInt64(&s.flagged, 1) }
func (s *Stats) IncrementBenign()  { atomic.AddInt64(&s.benign, 1) }
func (s *Stats) IncrementMalformed() {
	atomic.AddInt64(&s.malformed, 1)
}

func (s *Stats) IncrementInserted() { atomic.AddInt64(&s.inserted, 1) }
func (s *Stats) IncrementErrors()   { atomic.AddInt64(&s.errors, 1) }

func (s *Stats) IncrementGeoQueued()    { atomic.AddInt64(&s.geoQueued, 1) }
func (s *Stats) DecrementGeoQueued()    { atomic.AddInt64(&s.geoQueued, -1) }
func (s *Stats) IncrementGeoCompleted() { atomic.AddInt64(&s.geoCompleted, 1) }
func (s *Stats) IncrementGeoErrors()    { atomic.AddInt64(&s.geoErrors, 1) }

// Atomic getters
func (s *Stats) LinesRead() int64 { return atomic.LoadInt64(&s.linesRead) }
func (s *Stats) Flagged() int64   { return atomic.LoadInt64(&s.flagged) }
func (s *Stats) Benign() int64    { return atomic.LoadInt64(&s.benign) }
func (s *Stats) Malformed() int64 { return atomic.LoadInt64(&s.malformed) }
func (s *Stats) Inserted() int64  { return atomic.LoadInt64(&s.inserted) }
func (s *Stats) Errors() int64    { return atomic.LoadInt64(&s.errors) }

func (s *Stats) GeoQueued() int64    { return atomic.LoadInt64(&s.geoQueued) }
func (s *Stats) GeoCompleted() int64 { return atomic.LoadInt64(&s.geoCompleted) }
func (s *Stats) GeoErrors() int64    { return atomic.LoadInt64(&s.geoErrors) }

// Progress tracks how many lines have actually been “processed,” in addition to the Stats above.
type Progress struct {
	totalLines     int64
	processedLines int64 // lines that have made it through the DB insertion stage

	stats         *Stats
	messageBuffer []string

	lastDisplay time.Time
	startTime   time.Time

	// For computing rates over intervals
	lastTime         time.Time
	lastProcessed    int64
	lastGeoCompleted int64
}

// NewProgress creates a new Progress instance.
func NewProgress(totalLines int64, stats *Stats) *Progress {
	return &Progress{
		totalLines:    totalLines,
		stats:         stats,
		messageBuffer: make([]string, 0),
		lastDisplay:   time.Now(),
		startTime:     time.Now(),
		lastTime:      time.Now(),
	}
}

func (p *Progress) IncrementProcessed() {
	atomic.AddInt64(&p.processedLines, 1)
}
func (p *Progress) Processed() int64 {
	return atomic.LoadInt64(&p.processedLines)
}

// BufferMessage adds a message to be printed on the next display
func (p *Progress) BufferMessage(msg string) {
	p.messageBuffer = append(p.messageBuffer, msg)
}

// FlushMessages prints out all buffered messages
func (p *Progress) FlushMessages() {
	if len(p.messageBuffer) == 0 {
		return
	}
	fmt.Println("--- Messages ---")
	for _, m := range p.messageBuffer {
		fmt.Printf("  %s\n", m)
	}
	p.messageBuffer = p.messageBuffer[:0]
	fmt.Println()
}

// DisplayIfNeeded only calls Display() if minInterval has elapsed since the last display.
func (p *Progress) DisplayIfNeeded(minInterval time.Duration) {
	now := time.Now()
	if now.Sub(p.lastDisplay) >= minInterval {
		p.Display()
		p.lastDisplay = now
	}
}

// Display prints a one-time summary of current stats, plus “rate since last display.”
func (p *Progress) Display() {
	p.FlushMessages()

	now := time.Now()
	elapsedSinceLast := now.Sub(p.lastTime).Seconds()
	totalElapsed := now.Sub(p.startTime).Seconds()

	// Calculate deltas for rates
	curProcessed := p.Processed()
	curGeoCompleted := p.stats.GeoCompleted()

	linesDelta := curProcessed - p.lastProcessed
	geoDelta := curGeoCompleted - p.lastGeoCompleted

	linesRate := float64(linesDelta) / elapsedSinceLast
	geoRate := float64(geoDelta) / elapsedSinceLast

	// Save for next interval
	p.lastTime = now
	p.lastProcessed = curProcessed
	p.lastGeoCompleted = curGeoCompleted

	// Print a multi-line status block
	fmt.Printf("[%-24s] Elapsed: %6.2fs\n", now.Format("2006-01-02 15:04:05"), totalElapsed)

	fmt.Printf("  Lines:    read=%d    flagged=%d    benign=%d    malformed=%d\n",
		p.stats.LinesRead(), p.stats.Flagged(), p.stats.Benign(), p.stats.Malformed(),
	)
	fmt.Printf("  Processed: %d (DB inserted=%d)   Errors=%d\n",
		curProcessed, p.stats.Inserted(), p.stats.Errors(),
	)
	fmt.Printf("  Geo:       queued=%d   completed=%d   errors=%d\n",
		p.stats.GeoQueued(), curGeoCompleted, p.stats.GeoErrors(),
	)

	fmt.Printf("  Rates:     lines/s=%.2f   geo/s=%.2f\n", linesRate, geoRate)
	fmt.Println("---------------------------------------------------------------")
}

// StartPeriodicDisplay shows repeated progress updates until it receives from `done`.
func (p *Progress) StartPeriodicDisplay(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.Display()
		case <-done:
			// Final display & summary
			p.Display()
			p.finalSummary()
			return
		}
	}
}

// finalSummary prints a quick one-shot final summary after all processing is done
func (p *Progress) finalSummary() {
	fmt.Println("\n================= Final Summary =================")
	elapsed := time.Since(p.startTime).Seconds()
	fmt.Printf("Total Execution Time: %.2fs\n", elapsed)

	fmt.Printf("Lines Read:         %d\n", p.stats.LinesRead())
	fmt.Printf("Flagged (Threat):   %d\n", p.stats.Flagged())
	fmt.Printf("Benign:             %d\n", p.stats.Benign())
	fmt.Printf("Malformed:          %d\n", p.stats.Malformed())
	fmt.Printf("DB Inserted:        %d\n", p.stats.Inserted())
	fmt.Printf("Errors Encountered: %d\n", p.stats.Errors())

	fmt.Printf("Geo Lookups Queued:     %d\n", p.stats.GeoQueued()) // should be zero if fully processed
	fmt.Printf("Geo Lookups Completed:  %d\n", p.stats.GeoCompleted())
	fmt.Printf("Geo Lookup Errors:      %d\n", p.stats.GeoErrors())

	fmt.Printf("=================================================\n\n")
}
