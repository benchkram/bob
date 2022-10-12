package boblog

import (
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

// Progress tracks progress rendering at a certain interval on a new line
// the current percentage of bytes added from a total of maxBytes
//
// Inspired by https://github.com/schollz/progressbar
type Progress struct {
	// maxBytes is the total size of bytes that are tracked
	maxBytes int64
	// currentBytes is the total number of bytes tracked so far
	currentBytes int64

	// description is added as prefix on every render
	description string

	// currentPercent is percent of currentBytes from maxBytes
	currentPercent int
	// lastPercent is the last rendered percent
	lastPercent int

	// lock for Add operations
	lock sync.Mutex

	// lastRendered time
	lastRendered time.Time
	// intervalToRender sets an interval when to render the progress
	intervalToRender time.Duration
}

// NewProgress initialize a new progress ready to use
func NewProgress(maxBytes int64, description string, intervalToShow time.Duration) *Progress {
	return &Progress{
		maxBytes:         maxBytes,
		description:      description,
		lastRendered:     time.Now(),
		intervalToRender: intervalToShow,
	}
}

// Add will add the specified amount to the progressbar
func (p *Progress) Add(num int) {
	p.Add64(int64(num))
}

// Add64 will add the specified amount to the progressbar
func (p *Progress) Add64(num int64) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.currentBytes += num
	p.currentPercent = int(float64(p.currentBytes) / float64(p.maxBytes) * 100)

	if p.currentPercent == p.lastPercent {
		return
	}

	if p.currentBytes == p.maxBytes || time.Since(p.lastRendered) >= p.intervalToRender {
		p.render()
	}
}

// render current progress ex. `description 54% (7.4kB/7.4kB)`
func (p *Progress) render() {
	currentHuman, currentUnit := humanizeBytes(float64(p.currentBytes))
	maxHuman, maxUnit := humanizeBytes(float64(p.maxBytes))

	fmt.Fprintf(os.Stdout, "%s %d%% (%s%s/%s%s)\n", p.description, p.currentPercent, currentHuman, currentUnit, maxHuman, maxUnit)
	p.lastRendered = time.Now()
	p.lastPercent = p.currentPercent
}

// Finish sets the current progress to 100%
func (p *Progress) Finish() {
	p.lock.Lock()
	p.currentBytes = p.maxBytes
	p.lock.Unlock()
	p.Add(0)
}

// Reader will wrap an io.Reader adding progress functionality
type Reader struct {
	io.Reader
	bar *Progress
}

// NewReader return a new Reader with a given progress
func NewReader(r io.Reader, bar *Progress) Reader {
	return Reader{
		Reader: r,
		bar:    bar,
	}
}

// Read will read n bytes and track progress
func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.bar.Add(n)
	return n, err
}

// Close the reader when it implements io.Closer
func (r *Reader) Close() (err error) {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	r.bar.Finish()
	return
}

func humanizeBytes(s float64) (string, string) {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	base := 1024.0
	if s < 10 {
		return fmt.Sprintf("%2.0f", s), sizes[0]
	}
	e := math.Floor(logn(s, base))
	suffix := sizes[int(e)]
	val := math.Floor(s/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f, val), suffix
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
