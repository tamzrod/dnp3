package panels

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/session"
)

// ProtocolPanel displays decoded protocol information.
type ProtocolPanel struct {
	container *fyne.Container

	dllLabel *widget.Label
	tlLabel  *widget.Label
	alLabel  *widget.Label
}

// NewProtocolPanel creates a new protocol panel.
func NewProtocolPanel() *ProtocolPanel {
	p := &ProtocolPanel{}

	title := widget.NewLabel("PROTOCOL DECODER")
	title.TextStyle.Bold = true

	p.dllLabel = widget.NewLabel("DLL: --")
	p.tlLabel = widget.NewLabel("TL:  --")
	p.alLabel = widget.NewLabel("AL:  --")

	// Style labels for code/monospace appearance
	p.dllLabel.TextStyle.Monospace = true
	p.tlLabel.TextStyle.Monospace = true
	p.alLabel.TextStyle.Monospace = true

	p.container = container.NewVBox(
		title,
		layout.NewSpacer().MinSize().Height,
		p.dllLabel,
		p.tlLabel,
		p.alLabel,
	)

	return p
}

// Container returns the panel container.
func (p *ProtocolPanel) Container() *fyne.Container {
	return p.container
}

// Update updates the panel with decoded protocol information.
func (p *ProtocolPanel) Update(resp *session.Response) {
	if resp == nil {
		return
	}

	// Format DLL layer (mock for MVP)
	p.dllLabel.SetText("DLL: DIR=0 PRM=1 FCB=1 FCV=1 FUNC=3 DEST=1024 SRC=1 LEN=12")

	// Format TL layer (mock for MVP)
	p.tlLabel.SetText("TL:  FIR=1 FIN=1 CON=0 UNS=0 SEQ=5")

	// Format AL layer based on response data
	var alText strings.Builder
	alText.WriteString("AL:  FUNC=RESPONSE (0x81)")

	if len(resp.BinaryInputs) > 0 {
		alText.WriteString(fmt.Sprintf(" | Objects: Group 1 Var 1 (%d points)", len(resp.BinaryInputs)))
	}
	if len(resp.AnalogInputs) > 0 {
		alText.WriteString(fmt.Sprintf(" | Objects: Group 30 Var 1 (%d points)", len(resp.AnalogInputs)))
	}
	if len(resp.Counters) > 0 {
		alText.WriteString(fmt.Sprintf(" | Objects: Group 20 Var 1 (%d points)", len(resp.Counters)))
	}

	p.alLabel.SetText(alText.String())
}

// Clear clears the protocol display.
func (p *ProtocolPanel) Clear() {
	p.dllLabel.SetText("DLL: --")
	p.tlLabel.SetText("TL:  --")
	p.alLabel.SetText("AL:  --")
}
