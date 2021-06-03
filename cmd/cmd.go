package cmd

import (
	"fmt"
	"go/build"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"git.chinawayltd.com/golib/gin-swagger/swagger"
	"github.com/spf13/cobra"
)

var (
	packageName           string
	doTrace               bool
	doPProfCPU            bool
	doPProfMem            bool
	swaggerOutputFileName string
)

func getPackageName() string {
	pwd, _ := os.Getwd()
	pkg, err := build.ImportDir(pwd, build.FindOnly)
	if err != nil {
		panic(err)
	}
	return pkg.ImportPath
}

var cmdRoot = &cobra.Command{
	Use:   "gin-swagger",
	Short: "Generate swagger.json from gin framework codes",
	Run: func(cmd *cobra.Command, args []string) {
		nowString := time.Now().Format(time.RFC3339)
		// pprof cpu
		if doPProfCPU {
			cpuf, err := os.Create(nowString + ".cpu.prof")
			if err != nil {
				panic(err)
			}
			defer cpuf.Close()
			if err := pprof.StartCPUProfile(cpuf); err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()
		}
		//trace
		if doTrace {
			f, err := os.Create(time.Now().Format(time.RFC3339) + ".trace")
			if err != nil {
				panic(err)
			}
			defer f.Close()

			trace.Start(f)
			defer trace.Stop()
		}

		sc := swagger.NewScanner(packageName)
		sc.Output(swaggerOutputFileName)

		// pprof memory
		if doPProfMem {
			memf, err := os.Create(nowString + ".mem.prof")
			if err != nil {
				panic(err)
			}
			defer memf.Close()
			runtime.GC()
			if err := pprof.WriteHeapProfile(memf); err != nil {
				panic(err)
			}
		}
	},
}

func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cmdRoot.PersistentFlags().StringVarP(&packageName, "package", "p", getPackageName(), "package name for generating")
	cmdRoot.PersistentFlags().BoolVarP(&doTrace, "trace", "t", false, "do trace")
	cmdRoot.PersistentFlags().BoolVarP(&doPProfCPU, "cpuprof", "c", false, "do pprof of CPU")
	cmdRoot.PersistentFlags().BoolVarP(&doPProfMem, "memprof", "m", false, "do pprof of memory")
	cmdRoot.PersistentFlags().StringVarP(&swaggerOutputFileName, "swaggerfile", "s", "swagger.json", "file name of swagger output")
}
