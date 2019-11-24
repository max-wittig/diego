package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	statusStarted = "started"
	statusStopped = "stopped"
)

var (
	runningGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "running_containers_total",
		Help: "Currently running containers",
	})
	imagesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "running_containers",
		Help: "Running containers",
	}, []string{
		"name",
		"image",
		"createdAt",
	})
	containersTotalCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "containers_seen_total",
		Help: "The number of total containers that every ran",
	})
)

var trackedContainers []runningContainer

type runningContainer struct {
	ID        string
	Command   string
	Image     string
	CreatedAt string
	Names     string
	Status    string
}

func containerInList(list *[]runningContainer, container *runningContainer) bool {
	for _, c := range *list {
		if c.ID == container.ID {
			return true
		}
	}
	return false
}

func printStatus(c *runningContainer, prometheusEnabled bool, prometheusPort int, status string) {
	var au aurora.Value
	if status == statusStarted {
		au = aurora.Green(status)
	} else if status == statusStopped {
		au = aurora.Red(status)
	}
	log.Printf("%s - %s %s", aurora.Cyan(aurora.Bold(c.Names)), aurora.Cyan(c.Image), au.String())

	if prometheusEnabled {
		if status == statusStarted {
			runningGauge.Inc()
			containersTotalCounter.Inc()
			imagesGauge.WithLabelValues(c.Names, c.Image, c.CreatedAt).Inc()
		} else if status == statusStopped {
			runningGauge.Dec()
			imagesGauge.WithLabelValues(c.Names, c.Image, c.CreatedAt).Dec()
		}
	}
}

func trackContainers(containers *[]runningContainer, prometheusEnabled bool, prometheusPort int) bool {
	changed := false
	for _, c := range *containers {
		if !containerInList(&trackedContainers, &c) {
			trackedContainers = append(trackedContainers, c)
			printStatus(&c, prometheusEnabled, prometheusPort, statusStarted)
			changed = true
		}
	}
	for i, c := range trackedContainers {
		if !containerInList(containers, &c) {
			// container gone, delete from slice
			trackedContainers = append(trackedContainers[:i], trackedContainers[i+1:]...)
			printStatus(&c, prometheusEnabled, prometheusPort, statusStopped)
			changed = true
		}
	}

	return changed
}

func replaceAllInList(strs []string) []string {
	var cleaned []string
	for _, s := range strs {
		cleaned = append(cleaned, strings.ReplaceAll(s, "\"", ""))
	}
	return cleaned
}

func getContainers(executor string) []runningContainer {
	var psData []runningContainer
	if executor == "docker" {
		out, _ := exec.Command(executor, "ps", "--format", "\"{{.ID}} {{.Command}} {{.Image}} {{.Names}} {{.Status}} {{.CreatedAt}}\"").Output()
		scanner := bufio.NewScanner(strings.NewReader(string(out)))

		for scanner.Scan() {
			line := replaceAllInList(strings.Split(scanner.Text(), " "))
			c := runningContainer{
				ID:        line[0],
				Command:   line[1],
				Image:     line[2],
				Names:     line[3],
				Status:    line[4],
				CreatedAt: line[5],
			}
			psData = append(psData, c)
		}
	} else if executor == "podman" {
		out, err := exec.Command(executor, "ps", "--format", "json").Output()
		if err != nil {
			log.Fatalln("Failed to execute ps with your container executor")
		}

		err = json.Unmarshal(out, &psData)
		if err != nil {
			log.Fatalln("Podman send invalid json")
		}
	} else {
		log.Fatalln("Your container exector is not supported")
	}
	return psData
}

func poll(a *appOptions) {
	containers := getContainers(a.containerExecutor)
	trackContainers(&containers, a.prometheusEnabled, a.prometheusPort)
}

func runPrometheus(a *appOptions) {
	prometheus.MustRegister(imagesGauge)
	prometheus.MustRegister(runningGauge)
	prometheus.MustRegister(containersTotalCounter)
	_, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		panic("unexpected behavior of custom test registry")
	}
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Prometheus metrics server running on %d", a.prometheusPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", a.prometheusPort), nil))
}

func startWatching(a *appOptions) {
	if a.prometheusEnabled {
		go runPrometheus(a)
	}
	for true {
		poll(a)
		time.Sleep(time.Duration(a.interval) * time.Millisecond)
	}
}
