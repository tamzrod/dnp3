package tui

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Screen represents the terminal screen buffer.
type Screen struct {
	width  int
	height int
	buf    bytes.Buffer
}

// NewScreen creates a new screen with the given dimensions.
func NewScreen(width, height int) *Screen {
	return &Screen{
		width:  width,
		height: height,
	}
}

// Clear clears the entire screen.
func (s *Screen) Clear() {
	s.buf.WriteString(ClearScreen)
	s.buf.WriteString(MoveTo(1, 1))
}

// ClearLine clears the current line.
func (s *Screen) ClearLine() {
	s.buf.WriteString(ClearLine)
}

// Print writes text at the specified position (1-indexed).
func (s *Screen) Print(row, col int, text string) {
	s.buf.WriteString(MoveTo(row, col))
	s.buf.WriteString(text)
}

// PrintStyled writes styled text at the specified position.
func (s *Screen) PrintStyled(row, col int, text string, fg string, styles ...string) {
	s.buf.WriteString(MoveTo(row, col))
	s.buf.WriteString(Style(fg, styles...))
	s.buf.WriteString(text)
	s.buf.WriteString(Reset)
}

// PrintBold writes bold text at the specified position.
func (s *Screen) PrintBold(row, col int, text string, fg string) {
	s.PrintStyled(row, col, text, fg, "bold")
}

// PrintReversed writes reversed (highlighted) text.
func (s *Screen) PrintReversed(row, col int, text string, fg, bg string) {
	s.buf.WriteString(MoveTo(row, col))
	s.buf.WriteString(Color(fg))
	s.buf.WriteString(BgColor(bg))
	s.buf.WriteString(Reverse)
	s.buf.WriteString(text)
	s.buf.WriteString(Reset)
}

// PrintLine prints a line of text at row, filling to screen width.
func (s *Screen) PrintLine(row int, text string, fg string) {
	s.buf.WriteString(MoveTo(row, 1))
	s.buf.WriteString(Style(fg))
	s.buf.WriteString(Truncate(text, s.width))
	s.buf.WriteString(Reset)
}

// PrintLinePadded prints a line with padding on the right.
func (s *Screen) PrintLinePadded(row int, text string, width int, fg string) {
	s.buf.WriteString(MoveTo(row, 1))
	s.buf.WriteString(Style(fg))
	s.buf.WriteString(PadRight(text, width))
	s.buf.WriteString(Reset)
}

// FillRect fills a rectangular area with a character.
func (s *Screen) FillRect(row, col, width, height int, char string, fg, bg string) {
	for r := row; r < row+height; r++ {
		s.buf.WriteString(MoveTo(r, col))
		s.buf.WriteString(Color(fg))
		s.buf.WriteString(BgColor(bg))
		s.buf.WriteString(strings.Repeat(char, width))
		s.buf.WriteString(Reset)
	}
}

// DrawBox draws a box with the given boundaries.
func (s *Screen) DrawBox(top, left, bottom, right int, title string, titleFg string) {
	width := right - left + 1

	// Top border
	s.buf.WriteString(MoveTo(top, left))
	s.buf.WriteString("┌")
	s.buf.WriteString(strings.Repeat("─", width-2))
	s.buf.WriteString("┐")

	// Bottom border
	s.buf.WriteString(MoveTo(bottom, left))
	s.buf.WriteString("└")
	s.buf.WriteString(strings.Repeat("─", width-2))
	s.buf.WriteString("┘")

	// Left and right borders
	for r := top + 1; r < bottom; r++ {
		s.buf.WriteString(MoveTo(r, left))
		s.buf.WriteString("│")
		s.buf.WriteString(MoveTo(r, right))
		s.buf.WriteString("│")
	}

	// Title
	if title != "" {
		titleLen := len(title) + 2
		titlePos := left + (width-titleLen)/2
		s.buf.WriteString(MoveTo(top, titlePos))
		s.buf.WriteString(Style(titleFg, "bold"))
		s.buf.WriteString(" " + title + " ")
		s.buf.WriteString(Reset)
	}
}

// DrawHeader draws a header line with text.
func (s *Screen) DrawHeader(row int, text string, fg, bg string) {
	s.buf.WriteString(MoveTo(row, 1))
	s.buf.WriteString(Color(fg))
	s.buf.WriteString(BgColor(bg))
	s.buf.WriteString(Bold)
	s.buf.WriteString(Truncate(text, s.width))
	s.buf.WriteString(Reset)
}

// DrawSeparator draws a horizontal separator line.
func (s *Screen) DrawSeparator(row int, chars string) {
	s.buf.WriteString(MoveTo(row, 1))
	s.buf.WriteString(Dim)
	s.buf.WriteString(strings.Repeat(chars, s.width/len(chars)))
	s.buf.WriteString(Reset)
}

// Flush outputs the buffer to stdout.
func (s *Screen) Flush() error {
	_, err := os.Stdout.Write(s.buf.Bytes())
	s.buf.Reset()
	return err
}

// Truncate truncates a string to the given width.
func Truncate(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width-3]) + "..."
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// PadRight pads a string with spaces to the given width.
func PadRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// Center centers text within the given width.
func Center(s string, width int) string {
	padding := (width - len(s)) / 2
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + s
}

// Filled returns a string of char repeated to fill width.
func Filled(char string, width int) string {
	return strings.Repeat(char, width/len(char)+1)[:width]
}

// Printf is like fmt.Printf but writes to the screen buffer.
func (s *Screen) Printf(row, col int, format string, args ...interface{}) {
	s.Print(row, col, fmt.Sprintf(format, args...))
}

// PrintStyledf is like Printf but with styling.
func (s *Screen) PrintStyledf(row, col int, format string, fg string, styles []string, args ...interface{}) {
	s.PrintStyled(row, col, fmt.Sprintf(format, args...), fg, styles...)
}
