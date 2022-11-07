package nix

import (
	"fmt"
	"time"
)

// buildProgress tracks building of a Nix package and write to std.Out its progress
// example output: `go_1_18: ....`
type buildProgress struct {
	// packageName is the name of the package being built ex. go_1_18
	packageName string
	// padding is the padding in spaces after pkg_name:
	padding string
	// start time of progress
	start time.Time

	// ticker used to write a `.` at a certain interval
	ticker *time.Ticker
	// done channel to signal end of progress tracking and exit from the started goroutine
	done chan bool
}

// newBuildProgress creates a new progress track for a package
func newBuildProgress(pkgName, padding string) *buildProgress {
	var bp buildProgress
	bp.packageName = pkgName
	bp.padding = padding
	bp.done = make(chan bool)
	return &bp
}

// Start will start progress tracking and write to os.Stout a dot after every duration passes
func (bp *buildProgress) Start(duration time.Duration) {
	fmt.Printf("%s:%s", bp.packageName, bp.padding)

	bp.ticker = time.NewTicker(duration)

	bp.start = time.Now()
	fmt.Print(".")

	go func() {
		for {
			select {
			case <-bp.done:
				return
			case <-bp.ticker.C:
				fmt.Print(".")
			}
		}
	}()
}

// Stop will stop the current progress output
func (bp *buildProgress) Stop() {
	bp.ticker.Stop()
	bp.done <- true
}

// Duration gives the duration of progress tracking
func (bp *buildProgress) Duration() time.Duration {
	return time.Since(bp.start)
}
