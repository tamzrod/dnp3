// Package panels provides UI panel components for the workbench.
package panels

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
)

// DataPanel displays DNP3 response data.
type DataPanel struct {
	container *fyne.Container

	binaryLabel  *widget.Label
	analogLabel  *widget.Label
	counterLabel *widget.Label
}

// NewDataPanel creates a new data panel.
func NewDataPanel() *DataPanel {
	p := &DataPanel{}

	title := widget.NewLabel("POINT VALUES")
	title.TextStyle.Bold = true

	p.binaryLabel = widget.NewLabel("Binary Inputs: (No data)")
	p.analogLabel = widget.NewLabel("Analog Inputs: (No data)")
	p.counterLabel = widget.NewLabel("Counters: (No data)")

	p.container = container.NewVBox(
		title,
		layout.NewSpacer().MinSize().Height,
		p.binaryLabel,
		p.analogLabel,
		p.counterLabel,
	)

	return p
}

// Container returns the panel container.
func (p *DataPanel) Container() *fyne.Container {
	return p.container
}

// Update updates the panel with new response data.
func (p *DataPanel) Update(resp *session.Response) {
	if resp == nil {
		return
	}

	// Update binary inputs
	if len(resp.BinaryInputs) > 0 {
		summary := fmt.Sprintf("Binary Inputs: %d points", len(resp.BinaryInputs))
		for i, bi := range resp.BinaryInputs {
			if i < 3 {
				value := "OFF"
				if bi.Value {
					value = "ON"
				}
				summary += fmt.Sprintf("\n  Index %d: %s", bi.Index, value)
			}
		}
		if len(resp.BinaryInputs) > 3 {
			summary += fmt.Sprintf("\n  ... and %d more", len(resp.BinaryInputs)-3)
		}
		p.binaryLabel.SetText(summary)
	} else {
		p.binaryLabel.SetText("Binary Inputs: (No data)")
	}

	// Update analog inputs
	if len(resp.AnalogInputs) > 0 {
		summary := fmt.Sprintf("Analog Inputs: %d points", len(resp.AnalogInputs))
		for i, ai := range resp.AnalogInputs {
			if i < 3 {
				summary += fmt.Sprintf("\n  Index %d: %.2f", ai.Index, ai.Value)
			}
		}
		if len(resp.AnalogInputs) > 3 {
			summary += fmt.Sprintf("\n  ... and %d more", len(resp.AnalogInputs)-3)
		}
		p.analogLabel.SetText(summary)
	} else {
		p.analogLabel.SetText("Analog Inputs: (No data)")
	}

	// Update counters
	if len(resp.Counters) > 0 {
		summary := fmt.Sprintf("Counters: %d points", len(resp.Counters))
		for i, c := range resp.Counters {
			if i < 3 {
				summary += fmt.Sprintf("\n  Index %d: %d", c.Index, c.Value)
			}
		}
		if len(resp.Counters) > 3 {
			summary += fmt.Sprintf("\n  ... and %d more", len(resp.Counters)-3)
		}
		p.counterLabel.SetText(summary)
	} else {
		p.counterLabel.SetText("Counters: (No data)")
	}
}

// Clear clears all data.
func (p *DataPanel) Clear() {
	p.binaryLabel.SetText("Binary Inputs: (No data)")
	p.analogLabel.SetText("Analog Inputs: (No data)")
	p.counterLabel.SetText("Counters: (No data)")
}
