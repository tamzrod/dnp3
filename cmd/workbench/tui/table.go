package tui

import (
	"strings"
)

// Column represents a table column.
type Column struct {
	Title    string
	Width    int
	Align    string // "left", "center", "right"
}

// Row represents a data row.
type Row struct {
	Cells []string
}

// Table represents a scrollable data table.
type Table struct {
	Columns    []Column
	Rows       []Row
	Cursor     int
	Offset     int
	Selected   int
	Bounds     Rect
	ShowCursor bool
}

// NewTable creates a new table.
func NewTable(bounds Rect) *Table {
	return &Table{
		Bounds:     bounds,
		ShowCursor: true,
		Columns:    []Column{},
		Rows:       []Row{},
	}
}

// SetColumns sets the table columns.
func (t *Table) SetColumns(cols []Column) {
	t.Columns = cols
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cells ...string) {
	t.Rows = append(t.Rows, Row{Cells: cells})
}

// SetRows replaces all rows.
func (t *Table) SetRows(rows []Row) {
	t.Rows = rows
	if t.Cursor >= len(rows) {
		t.Cursor = 0
	}
}

// SetRowsIfChanged replaces rows only if they differ from current rows.
// Returns true if rows were updated, false if no change.
func (t *Table) SetRowsIfChanged(rows []Row) bool {
	// Check if rows are the same length
	if len(rows) != len(t.Rows) {
		t.SetRows(rows)
		return true
	}
	
	// Check if any cell content changed
	for i := range rows {
		if !rowsEqual(rows[i], t.Rows[i]) {
			t.SetRows(rows)
			return true
		}
	}
	
	// No change needed
	return false
}

// rowsEqual compares two rows for equality.
func rowsEqual(a, b Row) bool {
	if len(a.Cells) != len(b.Cells) {
		return false
	}
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			return false
		}
	}
	return true
}

// Clear removes all rows.
func (t *Table) Clear() {
	t.Rows = []Row{}
	t.Cursor = 0
	t.Offset = 0
	t.Selected = -1
}

// MoveUp moves the cursor up.
func (t *Table) MoveUp() {
	if t.Cursor > 0 {
		t.Cursor--
		if t.Cursor < t.Offset {
			t.Offset = t.Cursor
		}
	}
}

// MoveDown moves the cursor down.
func (t *Table) MoveDown() {
	if t.Cursor < len(t.Rows)-1 {
		t.Cursor++
		visibleHeight := t.Bounds.Bottom - t.Bounds.Top - 2 // -2 for header
		if t.Cursor >= t.Offset+visibleHeight {
			t.Offset = t.Cursor - visibleHeight + 1
		}
	}
}

// Select selects the current row.
func (t *Table) Select() {
	if t.Cursor >= 0 && t.Cursor < len(t.Rows) {
		t.Selected = t.Cursor
	}
}

// Deselect deselects the current row.
func (t *Table) Deselect() {
	t.Selected = -1
}

// GetSelected returns the selected row index.
func (t *Table) GetSelected() int {
	return t.Selected
}

// Draw renders the table to the screen.
func (t *Table) Draw(s *Screen) {
	height := t.Bounds.Bottom - t.Bounds.Top + 1

	// Draw header
	s.Print(t.Bounds.Top, t.Bounds.Left, "┌")
	x := t.Bounds.Left + 1
	for i, col := range t.Columns {
		if i > 0 {
			s.Print(t.Bounds.Top, x, "│")
			x++
		}
		title := PadRight(col.Title, col.Width)
		s.PrintStyled(t.Bounds.Top, x, title, "cyan", "bold")
		x += col.Width
	}
	s.Print(t.Bounds.Top, x, "┐")

	// Draw separator
	s.Print(t.Bounds.Top+1, t.Bounds.Left, "├")
	x = t.Bounds.Top + 1
	for i, col := range t.Columns {
		if i > 0 {
			s.Print(x, t.Bounds.Left+sumWidths(t.Columns[:i])+i+1, "┼")
		}
		s.Print(x, t.Bounds.Left+sumWidths(t.Columns[:i+1])+i+2, "┤")
		for j := 0; j < col.Width; j++ {
			s.Print(x, t.Bounds.Left+sumWidths(t.Columns[:i])+i+2+j, "─")
		}
	}

	// Draw rows
	visibleRows := height - 3 // header + separator + bottom
	for i := 0; i < visibleRows && t.Offset+i < len(t.Rows); i++ {
		row := t.Rows[t.Offset+i]
		rowNum := t.Offset + i
		y := t.Bounds.Top + 2 + i

		// Determine if this row is selected
		isSelected := rowNum == t.Selected
		isCursor := rowNum == t.Cursor && t.ShowCursor

		// Draw row
		s.Print(y, t.Bounds.Left, "│")
		x = t.Bounds.Left + 1
		for j, cell := range row.Cells {
			if j < len(t.Columns) {
				width := t.Columns[j].Width
				cellText := PadRight(cell, width)
				
				if isSelected {
					s.PrintStyled(y, x, cellText, "white", "reverse")
				} else if isCursor {
					s.PrintStyled(y, x, cellText, "yellow", "bold")
				} else {
					s.Print(y, x, cellText)
				}
				x += width
				if j < len(t.Columns)-1 {
					s.Print(y, x, "│")
					x++
				}
			}
		}
		s.Print(y, t.Bounds.Right, "│")
	}

	// Draw bottom border
	s.Print(t.Bounds.Top+height-1, t.Bounds.Left, "└")
	x = t.Bounds.Left + 1
	for i, col := range t.Columns {
		if i > 0 {
			s.Print(t.Bounds.Top+height-1, x, "┴")
			x++
		}
		for j := 0; j < col.Width; j++ {
			s.Print(t.Bounds.Top+height-1, x, "─")
			x++
		}
	}
	s.Print(t.Bounds.Top+height-1, x, "┘")

	// Draw scroll indicator if needed
	if len(t.Rows) > visibleRows {
		scrollY := t.Bounds.Top + 2 + (t.Offset * visibleRows / len(t.Rows))
		scrollHeight := visibleRows * visibleRows / len(t.Rows)
		if scrollHeight < 1 {
			scrollHeight = 1
		}
		s.PrintStyled(scrollY, t.Bounds.Right, "▐", "dim")
	}
}

// sumWidths calculates the sum of column widths.
func sumWidths(cols []Column) int {
	sum := 0
	for _, col := range cols {
		sum += col.Width
	}
	return sum
}

// DrawSimple renders a simple text table without box drawing.
func (t *Table) DrawSimple(s *Screen, startY int) {
	// Draw header
	header := ""
	for i, col := range t.Columns {
		if i > 0 {
			header += " │ "
		}
		header += PadRight(col.Title, col.Width)
	}
	s.PrintStyled(startY, 1, header, "cyan", "bold")

	// Draw separator
	sep := ""
	for i, col := range t.Columns {
		if i > 0 {
			sep += "─┼─"
		}
		sep += strings.Repeat("─", col.Width)
	}
	s.Print(startY+1, 1, sep)

	// Draw rows
	for i, row := range t.Rows {
		y := startY + 2 + i
		if y > t.Bounds.Bottom {
			break
		}
		
		line := ""
		for j, cell := range row.Cells {
			if j > 0 {
				line += " │ "
			}
			if j < len(t.Columns) {
				line += PadRight(cell, t.Columns[j].Width)
			}
		}
		
		if i == t.Selected {
			s.PrintStyled(y, 1, line, "white", "reverse")
		} else if i == t.Cursor {
			s.PrintStyled(y, 1, line, "yellow", "bold")
		} else {
			s.Print(y, 1, line)
		}
	}
}
