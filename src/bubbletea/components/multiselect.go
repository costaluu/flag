package components

import (
	"log"

	"github.com/charmbracelet/huh"
)

func FormMultiSelect(title string, options []huh.Option[string]) []string {
	var result []string = []string{}

	err := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Options(
				options...,
			).
			Title(title).
			Value(&result),
	)).WithTheme(huh.ThemeBase()).Run()

	if err != nil {
		log.Fatalln(err)
	}

	return result
}