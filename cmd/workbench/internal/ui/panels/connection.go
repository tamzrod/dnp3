package panels

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/controller"
)

// ConnectionPanel provides connection configuration and controls.
type ConnectionPanel struct {
	container    *fyne.Container
	addressEntry *widget.Entry
	portEntry    *widget.Entry
	connectBtn   *widget.Button
	disconnectBtn *widget.Button
	ctrl         *controller.Controller
	connected    bool

	OnConnect    func(address string, port int)
	OnDisconnect func()
}

// NewConnectionPanel creates a new connection panel.
func NewConnectionPanel(ctrl *controller.Controller) *ConnectionPanel {
	p := &ConnectionPanel{
		ctrl: ctrl,
	}

	title := widget.NewLabel("CONNECTION")
	title.TextStyle.Bold = true

	// IP Address
	addrLabel := widget.NewLabel("IP Address:")
	p.addressEntry = widget.NewEntry()
	p.addressEntry.SetText("localhost")

	// Port
	portLabel := widget.NewLabel("TCP Port:")
	p.portEntry = widget.NewEntry()
	p.portEntry.SetText("20000")

	// Buttons
	p.connectBtn = widget.NewButton("Connect", func() {
		if p.OnConnect != nil {
			address := p.addressEntry.Text
			port, err := strconv.Atoi(p.portEntry.Text)
			if err != nil {
				return
			}
			p.OnConnect(address, port)
		}
	})

	p.disconnectBtn = widget.NewButton("Disconnect", func() {
		if p.OnDisconnect != nil {
			p.OnDisconnect()
		}
	})
	p.disconnectBtn.Disable()

	buttonBox := container.NewHBox(p.connectBtn, p.disconnectBtn)

	p.container = container.NewVBox(
		title,
		layout.NewSpacer().MinSize().Height,
		addrLabel,
		p.addressEntry,
		portLabel,
		p.portEntry,
		buttonBox,
	)

	return p
}

// Container returns the panel container.
func (p *ConnectionPanel) Container() *fyne.Container {
	return p.container
}

// SetConnected updates the UI based on connection state.
func (p *ConnectionPanel) SetConnected(connected bool) {
	p.connected = connected
	if connected {
		p.connectBtn.Disable()
		p.disconnectBtn.Enable()
		p.addressEntry.Disable()
		p.portEntry.Disable()
	} else {
		p.connectBtn.Enable()
		p.disconnectBtn.Disable()
		p.addressEntry.Enable()
		p.portEntry.Enable()
	}
}

// IsConnected returns the current connection state.
func (p *ConnectionPanel) IsConnected() bool {
	return p.connected
}
