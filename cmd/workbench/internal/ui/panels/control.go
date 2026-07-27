// Package panels provides UI panel components for the workbench.
package panels

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SelectedPoint represents a selected output point for control.
type SelectedPoint struct {
	Type  PointType
	Index uint16
}

// ControlPanel provides controls for operating DNP3 outputs.
type ControlPanel struct {
	container *fyne.Container

	// Selected points
	selectedPoints map[uint16]*SelectedPoint
	mu             sync.RWMutex

	// UI components
	selectedList *widget.List
	pointLabel   *widget.Label
	doOnBtn      *widget.Button
	doOffBtn     *widget.Button
	aoInput      *widget.Entry
	aoSendBtn    *widget.Button
	selectToggle *widget.Check

	// Callbacks
	OnOperate func(pointType PointType, index uint16, value interface{})
}

// NewControlPanel creates a new control panel.
func NewControlPanel() *ControlPanel {
	p := &ControlPanel{
		selectedPoints: make(map[uint16]*SelectedPoint),
	}

	p.setupUI()
	return p
}

// setupUI creates the UI components.
func (p *ControlPanel) setupUI() {
	title := widget.NewLabel("CONTROL PANEL")
	title.TextStyle.Bold = true

	// Selected points label
	p.pointLabel = widget.NewLabel("No points selected")
	p.pointLabel.TextStyle.Italic = true

	// Select-Then-Operate toggle
	p.selectToggle = widget.NewCheck("Select-Then-Operate", func(checked bool) {
		// Enable SBO mode
	})
	p.selectToggle.SetChecked(true)

	// Binary Output controls
	doLabel := widget.NewLabel("Binary Output:")

	p.doOnBtn = widget.NewButtonWithIcon("ON", theme.ConfirmIcon(), func() {
		p.operateSelected(PointTypeDO, true)
	})
	p.doOnBtn.Importance = widget.HighImportance

	p.doOffBtn = widget.NewButtonWithIcon("OFF", theme.CancelIcon(), func() {
		p.operateSelected(PointTypeDO, false)
	})
	p.doOffBtn.Importance = widget.MediumImportance

	// Analog Output controls
	aoLabel := widget.NewLabel("Analog Output:")
	aoLabel.TextStyle.Italic = true

	p.aoInput = widget.NewEntry()
	p.aoInput.SetPlaceHolder("Enter value...")

	p.aoSendBtn = widget.NewButton("Set", func() {
		value := p.aoInput.Text
		if value == "" {
			return
		}
		p.operateSelectedValue(PointTypeAO, value)
	})
	p.aoSendBtn.Importance = widget.HighImportance

	// Create list for selected points
	p.selectedList = widget.NewList(
		func() int {
			p.mu.RLock()
			defer p.mu.RUnlock()
			return len(p.selectedPoints)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Point")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			p.mu.RLock()
			defer p.mu.RUnlock()
			i := 0
			for _, pt := range p.selectedPoints {
				if i == id {
					item.(*widget.Label).SetText(fmt.Sprintf("%s-%d", pt.Type.String(), pt.Index))
					return
				}
				i++
			}
		},
	)

	// Build layout
	controls := container.NewVBox(
		widget.NewLabel(""),
		doLabel,
		container.NewHBox(p.doOnBtn, p.doOffBtn),
		widget.NewLabel(""),
		aoLabel,
		container.NewHBox(p.aoInput, p.aoSendBtn),
	)

	// Left side: selected list
	leftPanel := container.NewVBox(
		widget.NewLabel("Selected Points:"),
		container.NewScroll(p.selectedList),
	)

	p.container = container.NewBorder(
		container.NewVBox(title, p.selectToggle),
		nil, nil, nil,
		container.NewHSplit(leftPanel, controls),
	)
}

// Container returns the panel container.
func (p *ControlPanel) Container() *fyne.Container {
	return p.container
}

// SelectPoint adds a point to the selection.
func (p *ControlPanel) SelectPoint(pointType PointType, index uint16) {
	// Only allow output types
	if pointType != PointTypeDO && pointType != PointTypeAO {
		return
	}

	key := index // Could also use pointType<<16 | index
	p.mu.Lock()
	p.selectedPoints[key] = &SelectedPoint{
		Type:  pointType,
		Index: index,
	}
	p.mu.Unlock()

	p.updateLabel()
	p.selectedList.Refresh()
}

// DeselectPoint removes a point from the selection.
func (p *ControlPanel) DeselectPoint(pointType PointType, index uint16) {
	key := index
	p.mu.Lock()
	delete(p.selectedPoints, key)
	p.mu.Unlock()

	p.updateLabel()
	p.selectedList.Refresh()
}

// ClearSelection clears all selected points.
func (p *ControlPanel) ClearSelection() {
	p.mu.Lock()
	p.selectedPoints = make(map[uint16]*SelectedPoint)
	p.mu.Unlock()

	p.updateLabel()
	p.selectedList.Refresh()
}

// operateSelected operates the selected points with a boolean value.
func (p *ControlPanel) operateSelected(pointType PointType, value bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, pt := range p.selectedPoints {
		if pt.Type == pointType && p.OnOperate != nil {
			p.OnOperate(pt.Type, pt.Index, value)
		}
	}
}

// operateSelectedValue operates the selected points with a string value.
func (p *ControlPanel) operateSelectedValue(pointType PointType, value string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, pt := range p.selectedPoints {
		if pt.Type == pointType && p.OnOperate != nil {
			p.OnOperate(pt.Type, pt.Index, value)
		}
	}
}

// updateLabel updates the selected points label.
func (p *ControlPanel) updateLabel() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := len(p.selectedPoints)
	if count == 0 {
		p.pointLabel.SetText("No points selected")
	} else if count == 1 {
		for _, pt := range p.selectedPoints {
			p.pointLabel.SetText(fmt.Sprintf("Selected: %s-%d", pt.Type.String(), pt.Index))
		}
	} else {
		p.pointLabel.SetText(fmt.Sprintf("Selected: %d points", count))
	}
}

// GetSelectedPoints returns the currently selected points.
func (p *ControlPanel) GetSelectedPoints() []*SelectedPoint {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*SelectedPoint, 0, len(p.selectedPoints))
	for _, pt := range p.selectedPoints {
		result = append(result, pt)
	}
	return result
}

// Enable enables/disables the control panel.
func (p *ControlPanel) Enable(enabled bool) {
	if enabled {
		p.doOnBtn.Enable()
		p.doOffBtn.Enable()
		p.aoInput.Enable()
		p.aoSendBtn.Enable()
	} else {
		p.doOnBtn.Disable()
		p.doOffBtn.Disable()
		p.aoInput.Disable()
		p.aoSendBtn.Disable()
	}
}
