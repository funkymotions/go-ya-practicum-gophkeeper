package primitive

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type saveBlockResultMsg struct {
	err error
}

func listenForSaveBlockResult(ctx context.Context, errChan chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return saveBlockResultMsg{err: context.Canceled}
		case err := <-errChan:
			return saveBlockResultMsg{err: err}
		}
	}
}
