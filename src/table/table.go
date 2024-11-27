package table

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// headers := []string{"FEATURE", "BLOCKS", "RECORDS", "LINKS"}
// 	data := [][]string{
// 		{"fttoooooooooooo", "DEV", "ON", "ON"},
// 		{"fast", "DEV", "ON", "ON"},
// 		{"s2i", "OFF", "OFF", "OFF"},
// 	}

func RenderTable(headers []string, rows [][]string) {
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	typeColors := map[string]lipgloss.Color{
		"ON": lipgloss.Color("42"),
		"OFF": lipgloss.Color("160"),
		"DEV": lipgloss.Color("27"),
		"STATE":  lipgloss.Color("27"),
		"FEATURE":  lipgloss.Color("42"),
		"ACTIVE": lipgloss.Color("42"),
		"NOT ACTIVE": lipgloss.Color("160"),
	}

	fullHeaders := []string{"#"}
	fullHeaders = append(fullHeaders, headers...)

	CapitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	var data [][]string = [][]string{}

	for i, row := range rows {
		index := []string{fmt.Sprintf("%d", i + 1)}

		data = append(data, append(index, row...))
	}

	renderTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(CapitalizeHeaders(fullHeaders)...).
		Rows(data...).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style = baseStyle
			style = style.Foreground(lipgloss.Color("252"))

			if row == 0 {
				style = headerStyle
			}

			if col == 0 {
				style = style.Width(3)
			}

			if col == 2 {
				if len(data[0]) == 6 {
					style = style.Width(10)
				} else {
					style = style.Width(7)
				}
			}

			switch col {
				case 1:
					style = style.Foreground(lipgloss.Color("220"))
				case 2, 3, 4:
					c := typeColors

					if row > 0 {
						color := c[fmt.Sprint(data[row-1][col])]
						style = style.Foreground(lipgloss.Color(color))
					}
			}

			return style
		})

	fmt.Println(renderTable)
}