package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
)

const (
	statusStarted = "started"
	statusStopped = "stopped"
)

var trackedContainers []runningContainer

type runningContainer struct {
	ID        string
	Command   string
	Image     string
	CreatedAt string
	StartedAt string
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

func printStatus(c *runningContainer, status string) {
	var au aurora.Value
	if status == statusStarted {
		au = aurora.Green(status)
	} else if status == statusStopped {
		au = aurora.Red(status)
	}
	log.Printf("%s - %s %s", aurora.Cyan(aurora.Bold(c.Names)), aurora.Cyan(c.Image), au.String())
}

func trackContainers(containers *[]runningContainer) bool {
	changed := false
	for _, c := range *containers {
		if !containerInList(&trackedContainers, &c) {
			trackedContainers = append(trackedContainers, c)
			printStatus(&c, statusStarted)
			changed = true
		}
	}
	for i, c := range trackedContainers {
		if !containerInList(containers, &c) {
			// container gone, delete from slice
			trackedContainers = append(trackedContainers[:i], trackedContainers[i+1:]...)
			printStatus(&c, statusStopped)
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
		out, _ := exec.Command(executor, "ps", "--format", "\"{{.ID}} {{.Command}} {{.Image}} {{.Names}} {{.Status}}\"").Output()
		scanner := bufio.NewScanner(strings.NewReader(string(out)))

		for scanner.Scan() {
			line := replaceAllInList(strings.Split(scanner.Text(), " "))
			c := runningContainer{ID: line[0], Command: line[1], Image: line[2], Names: line[3], Status: line[4]}
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
	trackContainers(&containers)
}

func startWatching(a *appOptions) {
	for true {
		poll(a)
		time.Sleep(time.Duration(a.interval) * time.Millisecond)
	}
}
