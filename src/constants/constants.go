package constants

import "github.com/charmbracelet/lipgloss"

var (
	APP_NAME = "flag"
	COMMAND = "flag"
	VERSION = "v0.0.4"
	MIN_FEATURE_CHARACTERS = 5
	ID_LENGTH = 25
)

var (
	AccentColor = "#f97900"
	MergeMark = "🔁"
	FileMark = "📄"
	CheckMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	XMark = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("⨯")
	InfoMark = lipgloss.NewStyle().Foreground(lipgloss.Color("31")).SetString("ⓘ")
	WarningMark = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).SetString("⚠")
)

var (
	STATE_ON = "ON"
	STATE_OFF = "OFF"
	STATE_DEV = "DEV"
)

var (
	FeatureFolder = ".features"
	WorkingTreeDirectory = "_workingtree"
)