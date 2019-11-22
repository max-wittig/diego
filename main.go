package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/max-wittig/diego/version"
)

type appOptions struct {
	containerExecutor string
	output            string
	interval          int
}

func parseOptions() (*appOptions, error) {
	versionFlag := flag.Bool("version", false, "Print current version")
	outputFlag := flag.String("output", "stdout", "Where to output to.")
	executorFlag := flag.String("executor", "docker", "Set executor to watch.")
	intervalFlag := flag.Int("interval", 1000, "Interval to watch in milliseconds, if watch supplied.")
	flag.Parse()
	aConfig := appOptions{}

	if *versionFlag {
		fmt.Println("Build Date:", version.BuildDate)
		fmt.Println("Git Commit:", version.GitCommit)
		fmt.Println("Version:", version.Version)
		fmt.Println("Go Version:", version.GoVersion)
		fmt.Println("OS / Arch:", version.OsArch)
		return nil, nil
	}

	aConfig.containerExecutor = *executorFlag
	aConfig.output = *outputFlag
	aConfig.interval = *intervalFlag

	return &aConfig, nil
}

func main() {
	appOptions, err := parseOptions()
	if err != nil {
		log.Fatalln(err)
	}
	startWatching(appOptions)
}
