package cmd

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

const logoRaw = `
██████╗ ███████╗ ██████╗ ██╗███████╗████████╗██████╗ ██╗   ██╗
██╔══██╗██╔════╝██╔════╝ ██║██╔════╝╚══██╔══╝██╔══██╗╚██╗ ██╔╝
██████╔╝█████╗  ██║  ███╗██║███████╗   ██║   ██████╔╝ ╚████╔╝
██╔══██╗██╔══╝  ██║   ██║██║╚════██║   ██║   ██╔══██╗  ╚██╔╝
██║  ██║███████╗╚██████╔╝██║███████║   ██║   ██║  ██║   ██║
╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝   ╚═╝
`

var (
	gradientStart = "#ff00ff"
	gradientEnd   = "#00ffff"
)

func renderLogo() string {
	lines := strings.Split(strings.TrimPrefix(logoRaw, "\n"), "\n")
	if len(lines) == 0 {
		return ""
	}

	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	startColor, _ := colorful.Hex(gradientStart)
	endColor, _ := colorful.Hex(gradientEnd)

	var result strings.Builder
	result.WriteString("\n")
	for _, line := range lines {
		for i, char := range line {
			if char == ' ' {
				result.WriteRune(char)
				continue
			}
			t := float64(i) / float64(maxWidth)
			c := startColor.BlendLuv(endColor, t)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex()))
			result.WriteString(style.Render(string(char)))
		}
		result.WriteString("\n")
	}

	return result.String()
}

var logo = renderLogo()
