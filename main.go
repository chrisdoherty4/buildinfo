package main

import (
	"debug/buildinfo"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"text/tabwriter"
	"time"

	"github.com/peterbourgon/ff/v3"
)

func main() {
	opts := parseFlags()

	bi, err := buildinfo.ReadFile(opts.Path)
	handleError(err)

	info := toBuildInfo(bi)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Build Info\n")
	fmt.Fprintf(w, "Path:\t%s\n", opts.Path)
	fmt.Fprintf(w, "Compiler:\t%s\n", info.GoVersion)
	fmt.Fprintf(w, "Main module:\t%s\n", info.Main.Path)
	fmt.Fprintf(w, "VCS:\t%s\n", info.VCS)
	fmt.Fprintf(w, "Revision:\t%s (%v)\n", info.Revision, info.Time)
	fmt.Fprintf(w, "Dirty:\t%t\n", info.Dirty)
	w.Flush()
}

type options struct {
	Modules bool
	Path    string
}

func parseFlags() options {
	fl := flag.NewFlagSet("buildinfo", flag.ExitOnError)
	help := fl.Bool("h", false, "display help")

	// TODO: add support for -m flag
	// modules := fl.Bool("m", false, "modules")
	modules := false

	err := ff.Parse(fl, os.Args)
	handleError(err)

	if *help {
		fmt.Printf("Usage: %s [options] <path>", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	// Arg 1 is the first non-flag argument.
	path := fl.Arg(1)
	if path == "" {
		fl.Usage()
		os.Exit(1)
	}

	path, err = filepath.Abs(path)
	handleError(err)

	return options{
		Modules: modules,
		Path:    path,
	}
}

type buildInfo struct {
	buildinfo.BuildInfo
	CompilerSettings []debug.BuildSetting
	VCS              string
	Revision         string
	Time             time.Time
	Dirty            bool
}

func toBuildInfo(info *buildinfo.BuildInfo) *buildInfo {
	bi := buildInfo{
		BuildInfo: *info,
	}

	for _, setting := range info.Settings {
		switch {
		case setting.Key == "vcs":
			bi.VCS = setting.Value
		case setting.Key == "vcs.revision":
			bi.Revision = setting.Value
		case setting.Key == "vcs.time":
			t, err := time.Parse(time.RFC3339, setting.Value)
			if err != nil {
				handleError(err)
			}
			bi.Time = t
		case setting.Key == "vcs.modified":
			bi.Dirty = (setting.Value == "true")
		default:
			bi.CompilerSettings = append(bi.CompilerSettings, setting)
		}
	}

	return &bi
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
