package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/max-wittig/diego/version"
)

type appOptions struct {
	containerExecutor string
	interval          int
	prometheusPort    int
	prometheusEnabled bool
}

func parseOptions() (*appOptions, error) {
	versionFlag := flag.Bool("version", false, "Print current version")
	prometheusPortFlag := flag.Int("prometheus-port", 8000, "Port to use for prometheus metrics")
	executorFlag := flag.String("executor", "docker", "Set executor to watch.")
	prometheusEnabledFlag := flag.Bool("prometheus", false, "Should the prometheus exporter server enabled")
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
	aConfig.interval = *intervalFlag
	aConfig.prometheusPort = *prometheusPortFlag
	aConfig.prometheusEnabled = *prometheusEnabledFlag

	return &aConfig, nil
}

func main() {
	appOptions, err := parseOptions()
	if err != nil {
		log.Fatalln(err)
	}
	startWatching(appOptions)
}
