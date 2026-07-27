// Package panels provides UI panel components for the workbench.
package panels

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
)

// PointType represents the type of a DNP3 data point.
type PointType int

const (
	PointTypeDI PointType = iota
	PointTypeDO
	PointTypeAI
	PointTypeAO
	PointTypeCounter
)

// String returns the string representation of the point type.
func (pt PointType) String() string {
	switch pt {
	case PointTypeDI:
		return "DI"
	case PointTypeDO:
		return "DO"
	case PointTypeAI:
		return "AI"
	case PointTypeAO:
		return "AO"
	case PointTypeCounter:
		return "CTR"
	default:
		return "?"
	}
}

// DataPoint represents a single data point for display.
type DataPoint struct {
	Index       uint16
	Type        PointType
	Value       string
	Quality     string
	Timestamp   string
	Selected    bool
	QualityGood bool
}

// DataTablePanel provides a table view of DNP3 data points.
type DataTablePanel struct {
	container *fyne.Container

	// Data
	points []DataPoint
	mu     sync.RWMutex

	// UI components
	table *widget.Table

	// Callbacks
	OnPointSelected func(pointType PointType, index uint16, selected bool)
	OnReadAll       func()
}

// NewDataTablePanel creates a new data table panel.
func NewDataTablePanel() *DataTablePanel {
	p := &DataTablePanel{
		points: make([]DataPoint, 0),
	}

	p.setupUI()
	return p
}

// setupUI creates the UI components.
func (p *DataTablePanel) setupUI() {
	title := widget.NewLabel("DATA MONITORING")
	title.TextStyle.Bold = true

	// Create table with proper template
	p.createTable()

	// Toolbar
	toolbar := p.createToolbar()

	// Wrap table in scroll container
	scrollContainer := container.NewScroll(p.table)
	scrollContainer.SetMinSize(fyne.NewSize(400, 300))

	// Main container
	p.container = container.NewBorder(
		container.NewVBox(title, toolbar),
		nil, nil, nil,
		scrollContainer,
	)
}

// createTable creates the data table.
func (p *DataTablePanel) createTable() {
	p.table = widget.NewTable(
		p.tableLength,
		func() fyne.CanvasObject { return widget.NewLabel("") },
		p.updateCell,
	)

	// Set column widths
	p.table.SetColumnWidth(0, 60)   // Index
	p.table.SetColumnWidth(1, 40)   // Type
	p.table.SetColumnWidth(2, 100)  // Value
	p.table.SetColumnWidth(3, 80)   // Quality
	p.table.SetColumnWidth(4, 150)  // Time

	// Show header row with column names
	p.table.ShowHeaderRow = true
}

// tableLength returns the number of rows and columns.
func (p *DataTablePanel) tableLength() (int, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// +1 for header row, 5 columns
	return len(p.points) + 1, 5
}

// updateCell updates a table cell with data.
func (p *DataTablePanel) updateCell(id widget.TableCellID, cell fyne.CanvasObject) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	label := cell.(*widget.Label)
	label.TextStyle = fyne.TextStyle{}

	headers := []string{"Index", "Type", "Value", "Quality", "Time"}

	if id.Row == 0 {
		// Header row
		if id.Col < len(headers) {
			label.SetText(headers[id.Col])
			label.TextStyle.Bold = true
		}
		return
	}

	// Data row
	row := id.Row - 1
	if row >= len(p.points) {
		label.SetText("")
		return
	}

	point := p.points[row]
	switch id.Col {
	case 0:
		label.SetText(fmt.Sprintf("%d", point.Index))
	case 1:
		label.SetText(point.Type.String())
	case 2:
		label.SetText(point.Value)
	case 3:
		label.SetText(point.Quality)
	case 4:
		label.SetText(point.Timestamp)
	}
}

// createToolbar creates the toolbar buttons.
func (p *DataTablePanel) createToolbar() *fyne.Container {
	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		p.Clear()
	})
	clearBtn.Importance = widget.MediumImportance

	readAllBtn := widget.NewButtonWithIcon("Read All", theme.ViewRefreshIcon(), func() {
		if p.OnReadAll != nil {
			p.OnReadAll()
		}
	})
	readAllBtn.Importance = widget.HighImportance

	return container.NewHBox(clearBtn, readAllBtn)
}

// Container returns the panel container.
func (p *DataTablePanel) Container() *fyne.Container {
	return p.container
}

// Update updates the table with new response data.
func (p *DataTablePanel) Update(resp *session.Response) {
	if resp == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.points = make([]DataPoint, 0)

	// Add binary inputs
	for _, bi := range resp.BinaryInputs {
		dp := DataPoint{
			Index:       bi.Index,
			Type:        PointTypeDI,
			Value:       boolToString(bi.Value),
			Quality:     bi.Quality.String(),
			QualityGood: bi.Quality.IsGood(),
		}
		if bi.Time != nil {
			dp.Timestamp = bi.Time.String()
		}
		p.points = append(p.points, dp)
	}

	// Add analog inputs
	for _, ai := range resp.AnalogInputs {
		dp := DataPoint{
			Index:       ai.Index,
			Type:        PointTypeAI,
			Value:       fmt.Sprintf("%.2f", ai.Value),
			Quality:     ai.Quality.String(),
			QualityGood: ai.Quality.IsGood(),
		}
		if ai.Time != nil {
			dp.Timestamp = ai.Time.String()
		}
		p.points = append(p.points, dp)
	}

	// Add counters
	for _, c := range resp.Counters {
		dp := DataPoint{
			Index:       c.Index,
			Type:        PointTypeCounter,
			Value:       fmt.Sprintf("%d", c.Value),
			Quality:     c.Quality.String(),
			QualityGood: c.Quality.IsGood(),
		}
		if c.Time != nil {
			dp.Timestamp = c.Time.String()
		}
		p.points = append(p.points, dp)
	}

	// Refresh table
	if p.table != nil {
		p.table.Refresh()
	}
}

// Clear clears all data points.
func (p *DataTablePanel) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.points = make([]DataPoint, 0)
	if p.table != nil {
		p.table.Refresh()
	}
}

// GetPoints returns all data points.
func (p *DataTablePanel) GetPoints() []DataPoint {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]DataPoint, len(p.points))
	copy(result, p.points)
	return result
}

// SetOnReadAll sets the callback for Read All button.
func (p *DataTablePanel) SetOnReadAll(callback func()) {
	p.OnReadAll = callback
}

// boolToString converts a boolean to a display string.
func boolToString(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
