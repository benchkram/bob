package cli

import (
	"os"
	"runtime/pprof"

	"github.com/benchkram/errz"
)

const _cpuprofile = "cpuprofile.prof"
const _memprofile = "memprofile.prof"

func profiling(cpuprofile, memprofile bool) func() {
	doOnStop := []func(){}
	stop := func() {
		for _, d := range doOnStop {
			if d != nil {
				d()
			}
		}
	}

	if cpuprofile {
		f, err := os.Create(_cpuprofile)
		errz.Fatal(err)

		_ = pprof.StartCPUProfile(f)
		doOnStop = append(doOnStop, pprof.StopCPUProfile)
	}

	if memprofile {
		f, err := os.Create(_memprofile)
		errz.Fatal(err)

		doOnStop = append(doOnStop, func() {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
		})
	}

	return stop
}
