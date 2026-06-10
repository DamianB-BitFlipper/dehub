package tui

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/dlvhdr/gh-dehub/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/context"
)

func (m *Model) openBrowser() tea.Cmd {
	taskId := fmt.Sprintf("open_browser_%d", time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    "Opening in browser",
		FinishedText: "Opened in browser",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	// Resolve the current row now, on the Update goroutine; the command
	// closure runs on a background goroutine where reading the model would
	// race with Update and could open whatever row is selected by then.
	var url string
	if currRow := m.getCurrRowData(); currRow != nil && !reflect.ValueOf(currRow).IsNil() {
		url = currRow.GetUrl()
	}
	openCmd := func() tea.Msg {
		if url == "" {
			return constants.TaskFinishedMsg{
				TaskId: taskId,
				Err:    errors.New("current selection doesn't have a URL"),
			}
		}
		b := browser.New("", os.Stdout, os.Stdin)
		err := b.Browse(url)
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}
	return tea.Batch(startCmd, openCmd)
}
