package tui

import "github.com/charmbracelet/lipgloss"

// Maahinen theme - Finnish folklore earth spirit colors
// Earthy, forest tones with mystical accents
var (
	// Primary colors - forest and earth
	ColorForestGreen = lipgloss.Color("#2d5a27")
	ColorMossGreen   = lipgloss.Color("#4a7c4e")
	ColorEarthBrown  = lipgloss.Color("#5c4033")
	ColorBarkBrown   = lipgloss.Color("#3d2914")
	ColorStonegrey   = lipgloss.Color("#6b6b6b")

	// Accent colors - mystical
	ColorMysticPurple = lipgloss.Color("#7b5ea7")
	ColorSpellBlue    = lipgloss.Color("#5e81ac")
	ColorRuneGold     = lipgloss.Color("#d4a574")
	ColorFrostSilver  = lipgloss.Color("#a3be8c")

	// UI colors
	ColorText       = lipgloss.Color("#d8dee9")
	ColorTextDim    = lipgloss.Color("#7f8c8d")
	ColorBorder     = lipgloss.Color("#4c566a")
	ColorBorderDim  = lipgloss.Color("#3b4252")
	ColorBackground = lipgloss.Color("#2e3440")
	ColorHighlight  = lipgloss.Color("#88c0d0")

	// Semantic colors
	ColorSuccess = lipgloss.Color("#a3be8c")
	ColorWarning = lipgloss.Color("#ebcb8b")
	ColorError   = lipgloss.Color("#bf616a")
	ColorInfo    = lipgloss.Color("#81a1c1")
)

// Panel styles
var (
	// Message history panel
	MessagePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	// Chat input area
	ChatInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMysticPurple).
			Padding(0, 1)

	ChatInputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight).
				Padding(0, 1)

	// Tool panel
	ToolPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	ToolPanelHiddenStyle = lipgloss.NewStyle()
)

// Message styles
var (
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true)

	UserLabelStyle = lipgloss.NewStyle().
			Foreground(ColorSpellBlue).
			Bold(true)

	AssistantLabelStyle = lipgloss.NewStyle().
				Foreground(ColorMysticPurple).
				Bold(true)

	AssistantMessageStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim).
				Italic(true)

	ToolMessageStyle = lipgloss.NewStyle().
				Foreground(ColorRuneGold)

	// Tool call one-liner styles for message history
	ToolCallPrefixStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")). // Yellow (ANSI)
				Bold(true)

	ToolCallOneLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")) // Gray (ANSI)

	ToolCallFailedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")) // Red (ANSI)

	ToolCallCancelledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")) // Dim gray (ANSI)
)

// Tool call styles
var (
	ToolCallStyle = lipgloss.NewStyle().
			Foreground(ColorRuneGold).
			Bold(true)

	ToolNameStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	ToolArgsStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim)

	ToolSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ToolErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	ToolPendingStyle = lipgloss.NewStyle().
				Foreground(ColorInfo)

	ToolRunningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	ToolCancelledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")) // Bright red (ANSI) for container compatibility
)

// Command autocomplete styles - use ANSI colors for container compatibility
var (
	CommandMenuStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMysticPurple).
				Padding(0, 1)

	// Use ANSI yellow for selected, gray for others (works in containers)
	CommandItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")). // Gray (ANSI)
				Padding(0, 1)

	CommandItemSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("11")). // Bright yellow (ANSI)
					Bold(true).
					Padding(0, 1)

	CommandDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")) // Gray (ANSI)
)

// Header/status styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorMysticPurple).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Background(ColorBorderDim).
			Padding(0, 1)

	ModelIndicatorStyle = lipgloss.NewStyle().
				Foreground(ColorFrostSilver)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")). // Bright yellow (ANSI) for container compatibility
			Bold(true)
)

// Confirmation dialog styles
var (
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorWarning).
			Padding(1, 2).
			Background(ColorBackground)

	// Simple dialog style for tool confirmation
	SimpleDialogStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMysticPurple).
				Padding(1, 2)

	DialogTitleStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	DialogButtonStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder)

	DialogButtonSelectedStyle = lipgloss.NewStyle().
					Foreground(ColorBackground).
					Background(ColorSuccess).
					Padding(0, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(ColorSuccess)

	// Simple yes/no confirmation styles - use ANSI colors for container compatibility
	ConfirmYesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Gray (ANSI)

	ConfirmYesSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")). // Bright green (ANSI)
				Bold(true)

	ConfirmNoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Gray (ANSI)

	ConfirmNoSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")). // Bright red (ANSI)
				Bold(true)
)

// Help text
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSpellBlue)

	AutoConfirmOnStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")). // Bright green (ANSI)
				Bold(true)

	ToolPanelOnStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")). // Bright cyan (ANSI)
				Bold(true)
)
