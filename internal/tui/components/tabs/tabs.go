package tabs

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type SectionTab struct {
	section section.Section
	spinner spinner.Model
}

type Model struct {
	sections         []section.Section
	sectionTabs      []SectionTab
	carousel         carousel.Model
	ctx              *context.ProgramContext
	latestVersion    string
	hasSearchSection bool
}

func NewModel(ctx *context.ProgramContext) Model {
	c := carousel.New(
		carousel.WithHeight(1),
		carousel.WithOverflowIndicators("←", "→"),
		carousel.WithSeparators(),
	)
	m := Model{
		carousel:         c,
		hasSearchSection: true,
	}
	m.UpdateProgramContext(ctx)

	return m
}

// SetHasSearchSection controls whether the tabs render an implicit search
// section at index 0 (special-cased on the right with a search icon).
// Views like Actions that have no global search bar should set this to false
// so their first section is rendered like any other tab.
func (m *Model) SetHasSearchSection(v bool) {
	m.hasSearchSection = v
}

func (m Model) Init() tea.Cmd {
	return m.fetchHasNewVersion()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case latestVersionMsg:
		m.latestVersion = msg.version
	case spinner.TickMsg:
		for i, tab := range m.sectionTabs {
			if tab.section.GetIsLoading() {
				var cmd tea.Cmd
				m.sectionTabs[i].spinner, cmd = tab.spinner.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	m.UpdateTabTitles()

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	c := m.viewSectionTabs()
	logo := m.viewLogo()
	content := m.ctx.Styles.Tabs.TabsRow.
		Width(m.ctx.ScreenWidth).
		Height(common.HeaderHeight).
		BorderBottom(false).
		Render(lipgloss.JoinHorizontal(lipgloss.Bottom,
			lipgloss.NewStyle().Width(
				m.ctx.ScreenWidth-lipgloss.Width(logo),
			).Render(c), logo))

	return lipgloss.JoinVertical(lipgloss.Left, content, m.focusDivider())
}

func (m Model) viewSectionTabs() string {
	if len(m.sectionTabs) == 0 {
		return ""
	}

	if !m.hasSearchSection {
		// No implicit search section; render every tab left-to-right.
		return m.renderSectionTabItems(0, len(m.sectionTabs))
	}

	left := m.renderSectionTabItems(1, len(m.sectionTabs))
	search := m.renderSectionTabItems(0, 1)
	spacing := strings.Repeat(" ", max(0, m.ctx.ScreenWidth-lipgloss.Width(left)-lipgloss.Width(search)))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, spacing, search)
}

func (m Model) renderSectionTabItems(start, end int) string {
	parts := make([]string, 0, max(0, end-start)*2)
	for i := start; i < end; i++ {
		if i > start {
			parts = append(parts, m.ctx.Styles.Tabs.TabSeparator.Render("|"))
		}
		style := m.ctx.Styles.Tabs.Tab
		if m.carousel.Cursor() == i {
			style = m.ctx.Styles.Tabs.ActiveTab
		}
		parts = append(parts, style.Render(m.sectionTabTitle(i)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m Model) sectionTabTitle(i int) string {
	if i < 0 || i >= len(m.sectionTabs) {
		return ""
	}
	cfg := m.sectionTabs[i].section.GetConfig()
	title := cfg.Title
	isSearchIndex := m.hasSearchSection && i == 0
	if isSearchIndex {
		if title == "" {
			title = constants.SearchIcon
		}
	} else if m.sectionTabs[i].section.GetIsLoading() {
		title = fmt.Sprintf("%s %s", title, m.sectionTabs[i].spinner.View())
	} else if m.ctx.Config.Theme.Ui.SectionsShowCount {
		title = fmt.Sprintf("%s (%s)", title, utils.ShortNumber(m.sectionTabs[i].section.GetTotalCount()))
	}
	return title
}

func (m Model) focusDivider() string {
	primary := color.Color(m.ctx.Theme.PrimaryBorder)
	focus := color.Color(lipgloss.Color("#F6E58D"))
	line := strings.Repeat("━", max(0, m.ctx.ScreenWidth))
	if !m.ctx.SidebarOpen || m.ctx.PreviewPosition == "bottom" {
		color := primary
		if m.ctx.ActivePane == "main" {
			color = focus
		}
		return lipgloss.NewStyle().Foreground(color).Render(line)
	}

	mainWidth := max(0, min(m.ctx.MainContentWidth, m.ctx.ScreenWidth))
	previewWidth := max(0, m.ctx.ScreenWidth-mainWidth)
	mainColor := primary
	previewColor := primary
	if m.ctx.ActivePane == "preview" {
		previewColor = focus
	} else {
		mainColor = focus
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(mainColor).Render(strings.Repeat("━", mainWidth)),
		lipgloss.NewStyle().Foreground(previewColor).Render(strings.Repeat("━", previewWidth)),
	)
}

type latestVersionMsg struct {
	version string
	err     error
}

func (m *Model) fetchHasNewVersion() tea.Cmd {
	return func() tea.Msg {
		r, err := data.FetchLatestVersion()
		return latestVersionMsg{
			version: r.Repository.LatestRelease.TagName,
			err:     err,
		}
	}
}

func (m *Model) CurrSectionId() int {
	return m.carousel.Cursor()
}

func (m *Model) SetCurrSectionId(id int) {
	m.carousel.SetCursor(id)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.carousel.SetStyles(carousel.Styles{
		Item:              ctx.Styles.Tabs.Tab,
		Selected:          ctx.Styles.Tabs.ActiveTab,
		OverflowIndicator: ctx.Styles.Tabs.OverflowIndicator,
		Separator:         ctx.Styles.Tabs.TabSeparator,
	})

	m.carousel.SetWidth(ctx.ScreenWidth - lipgloss.Width(m.viewLogo()))
}

func (m *Model) SetSections(sections []section.Section) {
	sectionTabs := make([]SectionTab, 0)
	for _, s := range sections {
		tab := SectionTab{section: s, spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot), spinner.WithStyle(
				lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).PaddingLeft(2),
			),
		)}
		sectionTabs = append(sectionTabs, tab)
	}
	m.sectionTabs = sectionTabs
	m.UpdateTabTitles()
}

func (m *Model) UpdateTabTitles() {
	titles := make([]string, 0)
	for i, tab := range m.sectionTabs {
		cfg := tab.section.GetConfig()
		title := cfg.Title
		isSearchIndex := m.hasSearchSection && i == 0
		if isSearchIndex {
			if title == "" {
				title = constants.SearchIcon
			}
		} else if tab.section.GetIsLoading() {
			title = fmt.Sprintf("%s %s", title, m.sectionTabs[i].spinner.View())
		} else if m.ctx.Config.Theme.Ui.SectionsShowCount {
			title = fmt.Sprintf("%s (%s)", title,
				utils.ShortNumber(tab.section.GetTotalCount()))
		}

		titles = append(titles, title)
	}

	oldCursor := m.carousel.Cursor()
	m.carousel.SetItems(titles)
	m.carousel.SetCursor(oldCursor)
}

func (m *Model) viewLogo() string {
	version := lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(m.ctx.Version)
	if m.latestVersion != "" && m.ctx.Version != "dev" && m.ctx.Version != m.latestVersion {
		version = lipgloss.JoinVertical(
			lipgloss.Left,
			version,
			lipgloss.NewStyle().
				Foreground(m.ctx.Styles.Colors.SuccessText).
				Render(" Update available!"),
		)
	} else {
		version = lipgloss.PlaceVertical(2, lipgloss.Bottom, version)
	}

	return lipgloss.NewStyle().
		Padding(0, 1, 0, 2).
		Height(2).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Bottom,
			lipgloss.NewStyle().Foreground(context.LogoColor).Render(constants.Logo),
			" ",
			version,
		))
}

func (m *Model) SetAllLoading() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.sectionTabs {
		cmds = append(cmds, m.sectionTabs[i].spinner.Tick)
	}

	return cmds
}
