package daemon_tracker

import (
	"errors"
	"fmt"
	"os/exec"
)

type DaemonTracker struct {
	startedDaemons   []string
	stoppedDaemons   []string
	restartedDaemons []string
	enabledDaemons   []string
	disabledDaemons  []string
}

func NewDaemonTracker() DaemonTracker {
	return DaemonTracker{
		startedDaemons:   make([]string, 0),
		stoppedDaemons:   make([]string, 0),
		restartedDaemons: make([]string, 0),
	}
}

func (daemonTracker *DaemonTracker) StartDaemon(daemonName string) error {
	daemonTracker.startedDaemons = append(daemonTracker.startedDaemons, daemonName)

	conflict := daemonTracker.FindConflicts()
	if conflict != nil {
		return conflict
	}

	return runSystemdCommand("start", daemonName)
}

func (daemonTracker *DaemonTracker) StopDaemon(daemonName string) error {
	daemonTracker.stoppedDaemons = append(daemonTracker.stoppedDaemons, daemonName)

	conflict := daemonTracker.FindConflicts()
	if conflict != nil {
		return conflict
	}

	return runSystemdCommand("stop", daemonName)
}

func (daemonTracker *DaemonTracker) RestartDaemon(daemonName string) error {
	daemonTracker.restartedDaemons = append(daemonTracker.restartedDaemons, daemonName)

	conflict := daemonTracker.FindConflicts()
	if conflict != nil {
		return conflict
	}

	return runSystemdCommand("restart", daemonName)
}

func (daemonTracker *DaemonTracker) EnableDaemon(daemonName string) error {
	daemonTracker.enabledDaemons = append(daemonTracker.enabledDaemons, daemonName)

	conflict := daemonTracker.FindConflicts()
	if conflict != nil {
		return conflict
	}

	return runSystemdCommand("enable", daemonName)
}

func (daemonTracker *DaemonTracker) DisableDaemon(daemonName string) error {
	daemonTracker.disabledDaemons = append(daemonTracker.disabledDaemons, daemonName)

	conflict := daemonTracker.FindConflicts()
	if conflict != nil {
		return conflict
	}

	return runSystemdCommand("disable", daemonName)
}

// FindConflicts returns a boolean indicating if there are any conflicts and a string with more details
func (daemonTracker *DaemonTracker) FindConflicts() error {
	haveMutual, mutualElement := haveMutualElement(daemonTracker.startedDaemons, daemonTracker.stoppedDaemons)
	if haveMutual {
		return fmt.Errorf("conflict: trying to start and stop at the same time %s daemon", mutualElement)
	}

	haveMutual, mutualElement = haveMutualElement(daemonTracker.restartedDaemons, daemonTracker.stoppedDaemons)
	if haveMutual {
		return fmt.Errorf("conflict: trying to restart and stop at the same time %s daemon", mutualElement)
	}

	haveMutual, mutualElement = haveMutualElement(daemonTracker.enabledDaemons, daemonTracker.disabledDaemons)
	if haveMutual {
		return fmt.Errorf("conflict: trying to enable and disable at the same time %s daemon", mutualElement)
	}

	return nil
}

// haveMutualElement returns bool and mutual element if exist
func haveMutualElement(list1, list2 []string) (bool, string) {
	// Create a map to store the elements of the first array
	elementMap := make(map[string]bool)

	// Populate the map with elements from the first array
	for _, element := range list1 {
		elementMap[element] = true
	}

	// Check if any element from the second array is present in the map
	for _, element := range list2 {
		if elementMap[element] {
			return true, element
		}
	}

	return false, ""
}

func runSystemdCommand(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.Output()
	if err != nil {
		outputMsgError := errors.New(string(output))
		return errors.Join(err, outputMsgError)
	}
	return nil
}
