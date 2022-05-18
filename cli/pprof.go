package cli

import (
	"os"
	"runtime/pprof"

	"github.com/benchkram/bob/pkg/boblog"
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
		boblog.Log.Info("cpu profile enabled")
		f, err := os.Create(_cpuprofile)
		errz.Fatal(err)

		err = pprof.StartCPUProfile(f)
		errz.Log(err)
		doOnStop = append(doOnStop, func() {
			pprof.StopCPUProfile()
			_ = f.Close()
			boblog.Log.Info("cpu profile stopped")
		})
	}

	if memprofile {
		boblog.Log.Info("memory profile enabled")
		f, err := os.Create(_memprofile)
		errz.Fatal(err)

		doOnStop = append(doOnStop, func() {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
			boblog.Log.Info("memory profile stopped")
		})
	}

	return stop
}
