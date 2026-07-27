// Package panels provides UI panel components for the workbench.
package panels

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
	points      []DataPoint
	mu          sync.RWMutex

	// UI components
	table         *widget.Table
	selectedIndex binding.Int

	// Callbacks
	OnPointSelected func(pointType PointType, index uint16, selected bool)
}

// NewDataTablePanel creates a new data table panel.
func NewDataTablePanel() *DataTablePanel {
	p := &DataTablePanel{
		points:        make([]DataPoint, 0),
		selectedIndex: binding.NewInt(),
	}

	p.setupUI()
	return p
}

// setupUI creates the UI components.
func (p *DataTablePanel) setupUI() {
	title := widget.NewLabel("DATA MONITORING")
	title.TextStyle.Bold = true

	// Create table
	p.createTable()

	// Toolbar
	toolbar := p.createToolbar()

	// Main container
	p.container = container.NewBorder(
		container.NewVBox(title, toolbar),
		nil, nil, nil,
		p.table,
	)
}

// createTable creates the data table.
func (p *DataTablePanel) createTable() {
	// Headers
	headers := []string{"Index", "Type", "Value", "Quality", "Time"}

	// Table dimensions
	colWidths := []float32{60, 40, 100, 70, 140}
	rowHeight := float32(30)

	p.table = widget.NewTable(
		p.tableLength,
		p.tableCreateCell,
		p.tableOnSelected,
	)
	p.table.SetColumnWidth(0, colWidths[0])
	p.table.SetColumnWidth(1, colWidths[1])
	p.table.SetColumnWidth(2, colWidths[2])
	p.table.SetColumnWidth(3, colWidths[3])
	p.table.SetColumnWidth(4, colWidths[4])
}

// tableLength returns the number of rows.
func (p *DataTablePanel) tableLength() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.points) + 1 // +1 for header
}

// tableCreateCell creates a table cell.
func (p *DataTablePanel) tableCreateCell(id widget.TableCellID, cell fyne.CanvasObject) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if id.Row == 0 {
		// Header row
		headers := []string{"Index", "Type", "Value", "Quality", "Time"}
		if id.Column < len(headers) {
			cell.(*widget.Label).SetText(headers[id.Column])
			cell.(*widget.Label).TextStyle.Bold = true
		}
		return
	}

	// Data row
	row := id.Row - 1
	if row >= len(p.points) {
		return
	}

	point := p.points[row]
	switch id.Column {
	case 0:
		cell.(*widget.Label).SetText(fmt.Sprintf("%d", point.Index))
	case 1:
		cell.(*widget.Label).SetText(point.Type.String())
	case 2:
		cell.(*widget.Label).SetText(point.Value)
	case 3:
		cell.(*widget.Label).SetText(point.Quality)
		if !point.QualityGood {
			cell.(*widget.Label).TextStyle.Color = theme.ErrorColor()
		}
	case 4:
		cell.(*widget.Label).SetText(point.Timestamp)
	}
}

// tableOnSelected handles cell selection.
func (p *DataTablePanel) tableOnSelected(id widget.TableCellID) {
	if id.Row == 0 {
		return // Ignore header row
	}

	p.mu.RLock()
	row := id.Row - 1
	if row >= len(p.points) {
		p.mu.RUnlock()
		return
	}
	point := p.points[row]
	p.mu.RUnlock()

	if p.OnPointSelected != nil {
		p.OnPointSelected(point.Type, point.Index, true)
	}
}

// createToolbar creates the toolbar buttons.
func (p *DataTablePanel) createToolbar() *fyne.Container {
	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		p.Clear()
	})

	readAllBtn := widget.NewButtonWithIcon("Read All", theme.ContentRefreshIcon(), func() {
		// Trigger read all callback if set
	})

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

	// Add binary outputs
	for _, bo := range resp.BinaryOutputs {
		dp := DataPoint{
			Index:       bo.Index,
			Type:        PointTypeDO,
			Value:       boolToString(bo.Value),
			Quality:     bo.Status.String(),
			QualityGood: true,
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

	// Add analog outputs
	for _, ao := range resp.AnalogOutputs {
		dp := DataPoint{
			Index:       ao.Index,
			Type:        PointTypeAO,
			Value:       fmt.Sprintf("%.2f", ao.Value),
			Quality:     ao.Status.String(),
			QualityGood: true,
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

// boolToString converts a boolean to a display string.
func boolToString(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
