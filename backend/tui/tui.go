package tui

import (
	"fmt"
	"math"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-stats/internal/cache"
)

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	purple  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c3aed"))
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("#06b6d4"))
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("#10b981"))
	amber   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))
	muted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b"))
	white   = lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e8f0"))
	bold    = lipgloss.NewStyle().Bold(true)

	box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2a2d3e")).
		Padding(0, 1)

	header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7c3aed")).
		MarginBottom(1)

	label = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8")).
		Width(18)

	kpiBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2a2d3e")).
		Padding(0, 2).
		Width(22)
)

// ─── Tick message ─────────────────────────────────────────────────────────────

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ─── Model ───────────────────────────────────────────────────────────────────

type Model struct {
	cache     *cache.DataCache
	port      int
	startTime time.Time
	width     int
	ticks     int
}

func New(c *cache.DataCache, port int) Model {
	return Model{
		cache:     c,
		port:      port,
		startTime: time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			go m.cache.Refresh()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tickMsg:
		m.ticks++
		return m, tick()
	}
	return m, nil
}

func (m Model) View() string {
	stats, sessions, projects, derived := m.cache.Get()

	uptime := time.Since(m.startTime).Round(time.Second)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	heapMB := float64(memStats.HeapAlloc) / 1024 / 1024

	// ── KPI row ─────────────────────────────────────────────────────────────
	totalSessions := 0
	totalMessages := 0
	if stats != nil {
		totalSessions = stats.TotalSessions
		totalMessages = stats.TotalMessages
	}
	// Use parsed sessions for cost (more accurate than cache)
	var totalCost float64
	for _, s := range sessions {
		totalCost += s.CostUSD
	}

	kpi := lipgloss.JoinHorizontal(lipgloss.Top,
		kpiBox.Render(
			muted.Render("Sessions")+"\n"+
				bold.Copy().Foreground(lipgloss.Color("#e2e8f0")).Render(fmt.Sprintf("%d", totalSessions))),
		" ",
		kpiBox.Render(
			muted.Render("Messages")+"\n"+
				bold.Copy().Foreground(lipgloss.Color("#e2e8f0")).Render(commaFmt(totalMessages))),
		" ",
		kpiBox.Render(
			muted.Render("Projects")+"\n"+
				bold.Copy().Foreground(lipgloss.Color("#06b6d4")).Render(fmt.Sprintf("%d", len(projects)))),
		" ",
		kpiBox.Render(
			muted.Render("Total Cost")+"\n"+
				bold.Copy().Foreground(lipgloss.Color("#10b981")).Render(fmtCost(totalCost))),
	)

	// ── Thinking speed box ───────────────────────────────────────────────────
	wpm := derived.ThinkingWPM
	wpmStr := "-"
	if wpm > 0 {
		wpmStr = fmt.Sprintf("%.0f wpm", wpm)
	}
	thinkBox := box.Width(60).Render(
		header.Render("⚡ Thinking Speed")+
			label.Render("Words / min:")+cyan.Bold(true).Render(wpmStr)+"\n"+
			label.Render("Think chars:")+muted.Render(humanNum(derived.ThinkingChars))+"\n"+
			label.Render("Avg response:")+muted.Render(func() string {
			if derived.ApiCallCount == 0 {
				return "-"
			}
			return fmtResponseTime(derived.ApiTimeMs / derived.ApiCallCount)
		}()),
	)

	// ── System box ───────────────────────────────────────────────────────────
	goroutines := runtime.NumGoroutine()
	lastRefresh := m.cache.LastRefresh()
	refreshStr := "loading…"
	if !lastRefresh.IsZero() {
		refreshStr = lastRefresh.Format("15:04:05")
	}
	sysBox := box.Width(60).Render(
		header.Render("🖥  System")+
			label.Render("Memory:")+amber.Render(fmt.Sprintf("%.1f MB", heapMB))+"\n"+
			label.Render("Goroutines:")+muted.Render(fmt.Sprintf("%d", goroutines))+"\n"+
			label.Render("Uptime:")+muted.Render(uptime.String())+"\n"+
			label.Render("Last refresh:")+muted.Render(refreshStr),
	)

	// ── Status bar ───────────────────────────────────────────────────────────
	readyStr := muted.Render("loading…")
	if m.cache.IsReady() {
		readyStr = green.Render("● ready")
	}
	statusBar := fmt.Sprintf("  %s  %s  %s",
		readyStr,
		muted.Render(fmt.Sprintf("http://localhost:%d", m.port)),
		muted.Render("[r] refresh  [q] quit"),
	)

	// ── Spinner ──────────────────────────────────────────────────────────────
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spin := purple.Render(spinners[m.ticks%len(spinners)])

	title := lipgloss.JoinHorizontal(lipgloss.Center,
		spin+" ",
		bold.Copy().Foreground(lipgloss.Color("#7c3aed")).Render("Claude Code Stats"),
		muted.Render(fmt.Sprintf("  ·  port %d", m.port)),
	)

	return "\n" +
		"  " + title + "\n\n" +
		indent(kpi, 2) + "\n\n" +
		indent(thinkBox, 2) + "\n" +
		indent(sysBox, 2) + "\n\n" +
		statusBar + "\n"
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func indent(s string, n int) string {
	return lipgloss.NewStyle().MarginLeft(n).Render(s)
}

func fmtCost(usd float64) string {
	if usd >= 100 {
		return fmt.Sprintf("$%.0f", usd)
	}
	if usd >= 1 {
		return fmt.Sprintf("$%.2f", usd)
	}
	return fmt.Sprintf("$%.4f", usd)
}

func commaFmt(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s,%03d", commaFmt(n/1000), n%1000)
}

func humanNum(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1e9)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1e3)
	}
	return fmt.Sprintf("%d", n)
}

func fmtResponseTime(ms int64) string {
	if ms <= 0 {
		return "—"
	}
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	if ms < 60_000 {
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}
	return fmt.Sprintf("%dm %ds", ms/60000, (ms%60000)/1000)
}

func fmtDuration(ms int64) string {
	if ms <= 0 {
		return "—"
	}
	d := time.Duration(ms) * time.Millisecond
	h := int(d.Hours())
	m := int(math.Mod(d.Minutes(), 60))
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
