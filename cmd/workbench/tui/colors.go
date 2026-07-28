// Package tui provides a terminal-based UI for the DNP3 workbench.
package tui

import "fmt"

// ANSI color and style codes for terminal output.
const (
	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// Text styles
	Bold           = "\033[1m"
	Dim            = "\033[2m"
	Italic         = "\033[3m"
	Underline      = "\033[4m"
	Blink          = "\033[5m"
	Reverse        = "\033[7m"
	Hidden         = "\033[8m"
	Strikethrough  = "\033[9m"

	// Reset
	Reset = "\033[0m"

	// Cursor control
	HideCursor    = "\033[?25l"
	ShowCursor    = "\033[?25h"
	SaveCursor    = "\033[s"
	RestoreCursor = "\033[u"
	ClearScreen   = "\033[2J"
	ClearLine     = "\033[K"
)

// MoveTo returns the ANSI escape sequence to move cursor to row, col (1-indexed).
func MoveTo(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

// Color returns foreground color code.
func Color(name string) string {
	switch name {
	case "black":
		return Black
	case "red":
		return Red
	case "green":
		return Green
	case "yellow":
		return Yellow
	case "blue":
		return Blue
	case "magenta":
		return Magenta
	case "cyan":
		return Cyan
	case "white":
		return White
	case "brightblack", "gray", "grey":
		return BrightBlack
	case "brightred":
		return BrightRed
	case "brightgreen":
		return BrightGreen
	case "brightyellow":
		return BrightYellow
	case "brightblue":
		return BrightBlue
	case "brightmagenta":
		return BrightMagenta
	case "brightcyan":
		return BrightCyan
	case "brightwhite":
		return BrightWhite
	default:
		return Reset
	}
}

// BgColor returns background color code.
func BgColor(name string) string {
	switch name {
	case "black":
		return BgBlack
	case "red":
		return BgRed
	case "green":
		return BgGreen
	case "yellow":
		return BgYellow
	case "blue":
		return BgBlue
	case "magenta":
		return BgMagenta
	case "cyan":
		return BgCyan
	case "white":
		return BgWhite
	default:
		return ""
	}
}

// Style combines foreground color and optional styles.
func Style(fg string, styles ...string) string {
	result := Color(fg)
	for _, s := range styles {
		switch s {
		case "bold":
			result += Bold
		case "dim":
			result += Dim
		case "italic":
			result += Italic
		case "underline":
			result += Underline
		case "blink":
			result += Blink
		case "reverse":
			result += Reverse
		case "hidden":
			result += Hidden
		}
	}
	return result
}
