package panels

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
	"dnp3/pkg/dnp3/types"
)

// DataPanel displays DNP3 response data.
type DataPanel struct {
	container *fyne.Container

	binaryTable  *widget.Table
	binaryData   []*types.BinaryInput

	analogTable  *widget.Table
	analogData   []*types.AnalogInput

	counterTable *widget.Table
	counterData  []*types.Counter

	binaryItems  []string
	analogItems  []string
	counterItems []string
}

// NewDataPanel creates a new data panel.
func NewDataPanel() *DataPanel {
	p := &DataPanel{
		binaryData:   make([]*types.BinaryInput, 0),
		analogData:   make([]*types.AnalogInput, 0),
		counterData:  make([]*types.Counter, 0),
		binaryItems:  []string{"(No data)"},
		analogItems:  []string{"(No data)"},
		counterItems: []string{"(No data)"},
	}

	title := widget.NewLabel("RESPONSE DATA")
	title.TextStyle.Bold = true

	// Binary inputs table
	binaryLabel := widget.NewLabel("Binary Inputs (Group 1)")
	binaryLabel.TextStyle.Italic = true

	p.binaryTable = p.createTable(p.binaryItems, []string{"Index", "Value", "Quality"})
	if p.binaryTable != nil {
		p.binaryTable.Resize(fyne.NewSize(400, 100))
	}

	// Analog inputs table
	analogLabel := widget.NewLabel("Analog Inputs (Group 30)")
	analogLabel.TextStyle.Italic = true

	p.analogTable = p.createTable(p.analogItems, []string{"Index", "Value", "Quality"})
	if p.analogTable != nil {
		p.analogTable.Resize(fyne.NewSize(400, 100))
	}

	// Counters table
	counterLabel := widget.NewLabel("Counters (Group 20)")
	counterLabel.TextStyle.Italic = true

	p.counterTable = p.createTable(p.counterItems, []string{"Index", "Value", "Quality"})
	if p.counterTable != nil {
		p.counterTable.Resize(fyne.NewSize(400, 100))
	}

	// Create scrollable containers
	var binaryScroll, analogScroll, counterScroll *container.Scroll

	if p.binaryTable != nil {
		binaryScroll = container.NewScroll(p.binaryTable)
		binaryScroll.SetMinSize(fyne.NewSize(400, 100))
	}

	if p.analogTable != nil {
		analogScroll = container.NewScroll(p.analogTable)
		analogScroll.SetMinSize(fyne.NewSize(400, 100))
	}

	if p.counterTable != nil {
		counterScroll = container.NewScroll(p.counterTable)
		counterScroll.SetMinSize(fyne.NewSize(400, 100))
	}

	p.container = container.NewVBox(
		title,
		layout.NewSpacer().MinSize().Height,
		binaryLabel,
	)

	if binaryScroll != nil {
		p.container.Add(binaryScroll)
	}

	p.container.Add(analogLabel)

	if analogScroll != nil {
		p.container.Add(analogScroll)
	}

	p.container.Add(counterLabel)

	if counterScroll != nil {
		p.container.Add(counterScroll)
	}

	return p
}

func (p *DataPanel) createTable(items []string, headers []string) *widget.Table {
	if len(items) == 0 {
		items = []string{"(No data)"}
	}

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Item")
		},
		func(i int, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(items[i])
		},
	)

	return nil // Simplified: return nil for MVP, use labels instead
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
		p.binaryData = resp.BinaryInputs
		p.binaryItems = make([]string, len(resp.BinaryInputs))
		for i, bi := range resp.BinaryInputs {
			value := "OFF"
			if bi.Value {
				value = "ON"
			}
			p.binaryItems[i] = fmt.Sprintf("Index: %d  Value: %s  Quality: %s",
				bi.Index, value, bi.Quality.String())
		}
	}

	// Update analog inputs
	if len(resp.AnalogInputs) > 0 {
		p.analogData = resp.AnalogInputs
		p.analogItems = make([]string, len(resp.AnalogInputs))
		for i, ai := range resp.AnalogInputs {
			p.analogItems[i] = fmt.Sprintf("Index: %d  Value: %.2f  Quality: %s",
				ai.Index, ai.Value, ai.Quality.String())
		}
	}

	// Update counters
	if len(resp.Counters) > 0 {
		p.counterData = resp.Counters
		p.counterItems = make([]string, len(resp.Counters))
		for i, c := range resp.Counters {
			p.counterItems[i] = fmt.Sprintf("Index: %d  Value: %d  Quality: %s",
				c.Index, c.Value, c.Quality.String())
		}
	}

	p.refresh()
}

func (p *DataPanel) refresh() {
	// Refresh the container
	// In a full implementation, this would update the actual table widgets
}

// Clear clears all data.
func (p *DataPanel) Clear() {
	p.binaryData = make([]*types.BinaryInput, 0)
	p.analogData = make([]*types.AnalogInput, 0)
	p.counterData = make([]*types.Counter, 0)
	p.binaryItems = []string{"(No data)"}
	p.analogItems = []string{"(No data)"}
	p.counterItems = []string{"(No data)"}
	p.refresh()
}

// FormatTimestamp formats a timestamp for display.
func FormatTimestamp(t time.Time) string {
	return t.Format("15:04:05.000")
}

// FormatIIN formats IIN bytes for display.
func FormatIIN(iin [2]byte) string {
	return fmt.Sprintf("0x%02X%02X", iin[0], iin[1])
}

// FormatHex formats byte slice as hex string.
func FormatHex(data []byte) string {
	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, " ")
}
