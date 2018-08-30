//go:generate go run $GOPATH/src/github.com/Al2Klimov/go-gen-source-repos/main.go github.com/Al2Klimov/check_linux_newkernel

package main

import (
	"errors"
	"fmt"
	_ "github.com/Al2Klimov/go-gen-source-repos"
	. "github.com/Al2Klimov/go-monplug-utils"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
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
var kernelInBoot = regexp.MustCompile(`\Avmlinuz`)

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

func checkLinuxNewkernel() (status uint8, output string, perfdata PerfdataCollection, errs map[string]error) {
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

	if len(krnels.kernels) < 1 {
		status = 1
		output = "No kernels found (ls /boot/vmlinuz*)"
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
			status = 0
			output = "No kernels have been installed since boot"
		} else {
			status = 2
			output = fmt.Sprintf("The kernel '/boot/%s' has been installed %s after boot", latestKernel, diff.String())
		}

		perfdata = PerfdataCollection{Perfdata{
			Label: "mtime_boot_diff",
			UOM:   "us",
			Value: float64(diff) / float64(time.Microsecond),
			Crit:  OptionalThreshold{IsSet: true, Inverted: true, Start: 0, End: math.Inf(1)},
		}}
	}

	return
}

func getBootTime(ch chan bootTime) {
	content, errRF := ioutil.ReadFile("/proc/uptime")
	if errRF != nil {
		ch <- bootTime{errs: map[string]error{"cat /proc/uptime": errRF}}
		return
	}

	now := time.Now()

	if match := procUptime.FindSubmatch(content); match == nil {
		ch <- bootTime{errs: map[string]error{"cat /proc/uptime": errors.New("bad output: " + string(content))}}
	} else if uptime, errPF := strconv.ParseFloat(string(match[1]), 64); errPF == nil {
		ch <- bootTime{bootTime: now.Add(time.Duration(-uptime)), errs: nil}
	} else {
		ch <- bootTime{errs: nil}
	}
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
