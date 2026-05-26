package actionview

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// itemMeta carries the per-row rendering state that is shared by every pane's
// list item (runs, jobs, checks, steps).
//
// Two distinct concepts are tracked separately on purpose:
//
//   - paneFocused: true iff the pane that owns this item currently has
//     keyboard focus. This is the literal "is this pane focused" signal and
//     should drive anything that genuinely depends on focus (cursor cues,
//     key hints, etc.).
//
//   - prominentSelection: true iff this pane's selected row should be drawn
//     with the bright "focused-selected" styling. This is a per-pane UX
//     policy that is intentionally allowed to diverge from paneFocused. For
//     example, the Checks and Steps panes opt to always keep their selection
//     prominent even when blurred, so the user can see which item produced
//     the logs they are currently reading. Runs and Jobs follow the
//     conventional "selection dims when blurred" pattern, which is set by
//     making prominentSelection track paneFocused.
//
// Keeping these as separate fields avoids overloading a single "focused"
// boolean with two meanings, which is what previously made the visual
// behavior of the Checks list look inconsistent with the Jobs/Runs lists.
type itemMeta struct {
	paneFocused        bool
	prominentSelection bool
	selected           bool
	styles             styles
	width              int
}

func (i itemMeta) TitleStyle() lipgloss.Style {
	if i.selected && i.prominentSelection {
		return i.styles.paneItem.focusedSelectedTitleStyle
	} else if i.selected {
		return i.styles.paneItem.selectedTitleStyle
	} else if i.prominentSelection {
		return i.styles.paneItem.focusedTitleStyle
	}

	return i.styles.paneItem.unfocusedTitleStyle
}

func (i itemMeta) DescStyle() lipgloss.Style {
	if i.selected && i.prominentSelection {
		w := i.width - i.styles.paneItem.focusedSelectedDescStyle.GetPaddingLeft() + 1
		return i.styles.paneItem.focusedSelectedDescStyle.Width(w).MaxHeight(1)
	} else if i.selected {
		w := i.width - i.styles.paneItem.selectedDescStyle.GetPaddingLeft() + 1
		return i.styles.paneItem.selectedDescStyle.Width(w).MaxHeight(1)
	}

	return i.styles.paneItem.descStyle.MaxHeight(1)
}

// renderTitleWithStatus is the shared title-rendering primitive used by
// runItem, jobItem and stepItem. It renders a status glyph, a one-space
// separator, and a truncated name, all styled through the item's current
// TitleStyle so selection/focus state is reflected consistently.
func (i itemMeta) renderTitleWithStatus(status, name string) string {
	s := i.TitleStyle()
	w := i.width - lipgloss.Width(status) - 2
	return lipgloss.JoinHorizontal(lipgloss.Top, s.Render(status), s.Render(" "),
		s.Width(w).Render(ansi.Truncate(s.Render(name), w, Ellipsis)))
}

// commonDelegate partially implements charm.land/bubbles.list.ItemDelegate.
//
// paneFocused and prominentSelection mirror the corresponding fields on
// itemMeta and are propagated into every item at render time. See itemMeta
// for the precise semantics.
type commonDelegate struct {
	paneFocused        bool
	prominentSelection bool
	styles             styles
}

func (d *commonDelegate) Render(
	w io.Writer,
	m list.Model,
	index int,
	item list.DefaultItem,
	meta *itemMeta,
) {
	isSelected := index == m.Index()
	meta.paneFocused = d.paneFocused
	meta.prominentSelection = d.prominentSelection
	meta.selected = isSelected
	meta.width = m.Width()

	var title, desc string

	title = item.Title()
	desc = item.Description()

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	itemStyle := lipgloss.NewStyle().PaddingLeft(1)
	if d.prominentSelection && isSelected {
		itemStyle = meta.styles.paneItem.focusedSelectedStyle
	} else if isSelected {
		itemStyle = meta.styles.paneItem.selectedStyle
	}

	textwidth := m.Width() - itemStyle.GetBorderLeftSize() - itemStyle.GetPaddingLeft()
	ts := meta.TitleStyle()
	title = ts.Render(title)
	ds := meta.DescStyle()
	desc = ds.Render(ansi.Truncate(desc, textwidth-ds.GetPaddingLeft(), Ellipsis))

	// TODO: implement filtering styles

	fmt.Fprintf(w, "%s", itemStyle.Width(m.Width()).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, desc),
	))
}

// Height implements charm.land/bubbles.list.ItemDelegate.Height
func (d *commonDelegate) Height() int {
	return 2
}

// Spacing implements charm.land/bubbles.list.ItemDelegate.Spacing
func (d *commonDelegate) Spacing() int {
	return 1
}
