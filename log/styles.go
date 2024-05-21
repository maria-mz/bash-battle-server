package log

import "github.com/charmbracelet/lipgloss"

var (
	mutedColor  = lipgloss.AdaptiveColor{Light: "#575279", Dark: "#e0def4"}
	redColor    = lipgloss.AdaptiveColor{Light: "#b4637a", Dark: "#eb6f92"}
	goldColor   = lipgloss.AdaptiveColor{Light: "#ea9d34", Dark: "#f6c177"}
	tealColor   = lipgloss.AdaptiveColor{Light: "#8DA101", Dark: "#3e8fb0"}
	purpleColor = lipgloss.AdaptiveColor{Light: "#907aa9", Dark: "#c4a7e7"}
)

type logStyles struct {
	DebugStyle lipgloss.Style
	InfoStyle  lipgloss.Style
	WarnStyle  lipgloss.Style
	ErrorStyle lipgloss.Style
	FatalStyle lipgloss.Style
}

func newLogStyles() logStyles {
	debugStyle := lipgloss.NewStyle().
		SetString("DEBUG").
		Foreground(mutedColor).
		Bold(true)

	infoStyle := lipgloss.NewStyle().
		SetString("INFO").
		Foreground(tealColor).
		Bold(true)

	warnStyle := lipgloss.NewStyle().
		SetString("WARN").
		Foreground(goldColor).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		SetString("ERROR").
		Foreground(redColor).
		Bold(true)

	fatalStyle := lipgloss.NewStyle().
		SetString("FATAL").
		Foreground(purpleColor).
		Bold(true)

	return logStyles{
		DebugStyle: debugStyle,
		InfoStyle:  infoStyle,
		WarnStyle:  warnStyle,
		ErrorStyle: errorStyle,
		FatalStyle: fatalStyle,
	}
}
