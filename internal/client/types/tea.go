package types

import tea "github.com/charmbracelet/bubbletea"

type NamedTeaModel interface {
	tea.Model
	GetTitle() string
	SetPrevModel(NamedTeaModel)
	IsAuthorizedModel() bool
}
