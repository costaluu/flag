package components

import (
	"github.com/charmbracelet/huh/spinner"
)

func FormSpinner(title string, runner func()) {
	_ = spinner.New().Title(title).Action(runner).Run()	
}