package components

import (
	"log"

	"github.com/charmbracelet/huh"
)

func FormConfirm(title string, affirmative string, negative string) bool {
	var result bool = false

	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(title).
			Affirmative(affirmative).
			Negative(negative).
			Value(&result).WithWidth(1),
	).WithWidth(5)).WithWidth(125).Run()

	if err != nil {
		log.Fatalln(err)
	}

	return result
}