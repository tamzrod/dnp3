package tui

// Layout defines the screen layout regions.
type Layout struct {
	Width      int
	Height     int
	HeaderSize int
	FooterSize int
}

// Rect represents a rectangular region.
type Rect struct {
	Top    int
	Left   int
	Bottom int
	Right  int
}

// NewLayout creates a new layout with the given dimensions.
func NewLayout(width, height int) *Layout {
	return &Layout{
		Width:      width,
		Height:     height,
		HeaderSize: 2, // Header + separator
		FooterSize: 2, // Controls + separator
	}
}

// Resize updates the layout for new dimensions.
func (l *Layout) Resize(width, height int) {
	l.Width = width
	l.Height = height
}

// HeaderBounds returns the header region.
func (l *Layout) HeaderBounds() Rect {
	return Rect{
		Top:    1,
		Left:   1,
		Bottom: l.HeaderSize - 1,
		Right:  l.Width,
	}
}

// MainBounds returns the main content region (between header and footer).
func (l *Layout) MainBounds() Rect {
	top := l.HeaderSize
	bottom := l.Height - l.FooterSize
	return Rect{
		Top:    top,
		Left:   1,
		Bottom: bottom,
		Right:  l.Width,
	}
}

// FooterBounds returns the footer region.
func (l *Layout) FooterBounds() Rect {
	return Rect{
		Top:    l.Height - l.FooterSize + 1,
		Left:   1,
		Bottom: l.Height,
		Right:  l.Width,
	}
}

// LogBounds returns the log panel region within main.
func (l *Layout) LogBounds() Rect {
	main := l.MainBounds()
	logHeight := 4 // 3 lines for log + 1 separator
	return Rect{
		Top:    main.Bottom - logHeight,
		Left:   main.Left,
		Bottom: main.Bottom,
		Right:  main.Right,
	}
}

// TableBounds returns the table region within main.
func (l *Layout) TableBounds() Rect {
	main := l.MainBounds()
	log := l.LogBounds()
	return Rect{
		Top:    main.Top,
		Left:   main.Left,
		Bottom: log.Top - 2, // 1 separator line
		Right:  main.Right,
	}
}

// ContentHeight returns the height of the main content area.
func (l *Layout) ContentHeight() int {
	return l.Height - l.HeaderSize - l.FooterSize
}

// TableHeight returns the height available for the table.
func (l *Layout) TableHeight() int {
	return l.ContentHeight() - 6 // 4 for log + 2 separators
}
