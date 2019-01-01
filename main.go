//go:generate go run vendor/github.com/Al2Klimov/go-gen-source-repos/main.go github.com/Al2Klimov/check_linux_newkernel

package main

import (
	"fmt"
	_ "github.com/Al2Klimov/go-gen-source-repos"
	linux "github.com/Al2Klimov/go-linux-apis"
	. "github.com/Al2Klimov/go-monplug-utils"
	pp "github.com/Al2Klimov/go-pretty-print"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
)

type bootTime struct {
	bootTime time.Time
	errs     map[string]error
}

type kernels struct {
	kernels map[string]time.Time
	errs    map[string]error
}

var procUptime = regexp.MustCompile(`\A(\d+(?:\.\d+)?) `)
var kernelInBoot = regexp.MustCompile(`\A(?:vmlinuz|kernel\.img\z)`)

func main() {
	os.Exit(ExecuteCheck(onTerminal, checkLinuxNewkernel))
}

func onTerminal() (output string) {
	return fmt.Sprintf(
		"For the terms of use, the source code and the authors\n"+
			"see the projects this program is assembled from:\n\n  %s\n",
		strings.Join(GithubcomAl2klimovGo_gen_source_repos, "\n  "),
	)
}

func checkLinuxNewkernel() (output string, perfdata PerfdataCollection, errs map[string]error) {
	chBootTime := make(chan bootTime, 1)
	chKernels := make(chan kernels, 1)

	go getBootTime(chBootTime)
	go getKernels(chKernels)

	btTime := <-chBootTime
	krnels := <-chKernels

	chBootTime = nil
	chKernels = nil

	if btTime.errs != nil {
		errs = btTime.errs
	}

	if krnels.errs != nil {
		if errs == nil {
			errs = krnels.errs
		} else {
			for context, err := range krnels.errs {
				errs[context] = err
			}
		}
	}

	if errs != nil {
		return
	}

	perfdata = PerfdataCollection{Perfdata{
		Label: "kernels",
		Value: float64(len(krnels.kernels)),
		Warn:  OptionalThreshold{true, false, 1, math.Inf(1)},
		Min:   OptionalNumber{true, 0},
	}}

	if len(krnels.kernels) < 1 {
		output = "No kernels found (ls /boot/{vmlinuz*,kernel.img})"
	} else {
		var latestKernel string
		var latestKernelMTime time.Time

		for kernel, mTime := range krnels.kernels {
			latestKernel = kernel
			latestKernelMTime = mTime
			break
		}

		for kernel, mTime := range krnels.kernels {
			if mTime.After(latestKernelMTime) {
				latestKernel = kernel
				latestKernelMTime = mTime
			}
		}

		diff := latestKernelMTime.Sub(btTime.bootTime)

		if diff < 0 {
			output = "No kernels have been installed since boot"
		} else {
			output = fmt.Sprintf("The kernel '/boot/%s' has been installed %s after boot", latestKernel, pp.Duration(diff))
		}

		perfdata = append(perfdata, Perfdata{
			Label: "mtime_boot_diff",
			UOM:   "us",
			Value: float64(diff) / float64(time.Microsecond),
			Crit:  OptionalThreshold{IsSet: true, Inverted: true, Start: 0, End: math.Inf(1)},
		})
	}

	return
}

func getBootTime(ch chan bootTime) {
	uptime, errGUT := linux.GetUptime()
	if errGUT != nil {
		ch <- bootTime{errs: map[string]error{"cat /proc/uptime": errGUT}}
		return
	}

	ch <- bootTime{bootTime: time.Now().Add(-uptime.UpTime), errs: nil}
}

func getKernels(ch chan kernels) {
	entries, errRD := ioutil.ReadDir("/boot")
	if errRD != nil {
		ch <- kernels{errs: map[string]error{"ls /boot": errRD}}
		return
	}

	krnels := map[string]time.Time{}

	for _, entry := range entries {
		if name := entry.Name(); kernelInBoot.MatchString(name) {
			krnels[name] = entry.ModTime()
		}
	}

	ch <- kernels{kernels: krnels, errs: nil}
}
