package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Action int

const (
	RestartService Action = iota
	ScaleService
)

func SelectAction() (Action, error) {
	options := []string{
		"Restart service",
		"Scale tasks",
	}

	selected, err := SelectOptionFromStrings("Select action", options)
	if err != nil {
		return 0, err
	}

	switch selected {
	case "Restart service":
		return RestartService, nil
	case "Scale tasks":
		return ScaleService, nil
	default:
		return 0, fmt.Errorf("unknown selection")
	}
}

// reuse simple selector
func SelectOptionFromStrings(label string, items []string) (string, error) {
	opts := make([]Option[string], len(items))

	for i, v := range items {
		opts[i] = Option[string]{
			Label: v,
			Data:  v,
		}
	}

	selected, err := SelectOption(opts)
	if err != nil {
		return "", err
	}

	return selected.Data, nil
}

func ConfirmRestart(service string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Are you sure you want to restart service '%s'? (y/yes): ", service)
	input, _ := reader.ReadString('\n')

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func PromptScale(current int32) (int32, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter desired number of tasks (0-4). Current is %d: ", current)

	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	input = strings.TrimSpace(input)

	val, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}

	if val < 0 || val > 4 {
		return 0, fmt.Errorf("must be between 0 and 4")
	}

	return int32(val), nil
}
