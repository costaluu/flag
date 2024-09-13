package components

import "github.com/charmbracelet/huh"

func FormInput(title string, validate func (value string) error) string {
	var value string

	huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(title).
			Prompt(">").
			Validate(validate).
			Value(&value),
	)).Run()

	return value
}