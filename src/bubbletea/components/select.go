package components

import (
	"log"

	"github.com/charmbracelet/huh"
)

func FormSelect(title string, options []huh.Option[string]) string {
	var result string

	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title(title).
			Options(
				options...
			).
			Value(&result),
	)).Run()
	
	if err != nil {
		log.Fatalln(err)
	}

	return result
}