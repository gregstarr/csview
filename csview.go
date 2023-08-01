package main

import (
	"csview/table"
	"csview/utils"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	fn := os.Args[1]
	records := utils.ReadCsv(fn)
	model := table.New(records)
	tm := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	if _, err := tm.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
