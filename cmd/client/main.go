package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client"
)

// TODO: add build info flag to retrieve metainfo
// var clientVersion = "n/a"
// var buildDate = "n/a"

func main() {
	model := client.New()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
