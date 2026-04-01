package ui

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

type Option[T any] struct {
	Label string
	Data  T
}

func SelectOption[T any](options []Option[T]) (Option[T], error) {
	if len(options) == 0 {
		return Option[T]{}, fmt.Errorf("no options to select from")
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▸ {{ .Label | cyan }}",
		Inactive: "  {{ .Label }}",
		Selected: "✔ {{ .Label | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select ECS Service",
		Items:     options,
		Templates: templates,
		Size:      10,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return Option[T]{}, err
	}

	return options[index], nil
}
