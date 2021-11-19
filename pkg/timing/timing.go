package timing

import (
	"fmt"
	"time"
)

func Track(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

type Measurement struct {
	// name of the measurement
	name string

	// sum is the total duration elapsed yet
	sum time.Duration

	// lastStart last time Start() was called
	lastStart time.Time

	// cycles is the number of Start()/Stop() cycles
	cycles uint
}

func NewMeasurement(name string) *Measurement {
	return &Measurement{
		name: name,
	}
}

func NewStartedMeasurement(name string) *Measurement {
	m := NewMeasurement(name)
	m.Start()
	return m
}

func (m *Measurement) Start() {
	m.lastStart = time.Now()
}

func (m *Measurement) Stop() {
	elapsed := time.Since(m.lastStart)

	m.sum = m.sum + elapsed
	m.cycles++
}

func (m *Measurement) String() string {
	return fmt.Sprintf("%s ran %d times for a total of %s", m.name, m.cycles, m.sum)
}
