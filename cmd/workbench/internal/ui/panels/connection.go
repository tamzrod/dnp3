package panels

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"dnp3/cmd/workbench/internal/controller"
)

// ConnectionPanel provides connection configuration and controls.
type ConnectionPanel struct {
	container     *fyne.Container
	addressEntry   *widget.Entry
	portEntry     *widget.Entry
	connectBtn    *widget.Button
	disconnectBtn *widget.Button
	ctrl          *controller.Controller
	connected     bool

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
	p.addressEntry.SetPlaceHolder("Enter IP address")

	// Port
	portLabel := widget.NewLabel("TCP Port:")
	p.portEntry = widget.NewEntry()
	p.portEntry.SetText("20000")
	p.portEntry.SetPlaceHolder("Enter port")

	// Buttons with icons (UX Standard Section 5.4)
	p.connectBtn = widget.NewButtonWithIcon("Connect", theme.ConfirmIcon(), func() {
		if p.OnConnect != nil {
			address := p.addressEntry.Text
			port, err := strconv.Atoi(p.portEntry.Text)
			if err != nil {
				return
			}
			p.OnConnect(address, port)
		}
	})
	p.connectBtn.Importance = widget.HighImportance

	p.disconnectBtn = widget.NewButtonWithIcon("Disconnect", theme.CancelIcon(), func() {
		if p.OnDisconnect != nil {
			p.OnDisconnect()
		}
	})
	p.disconnectBtn.Importance = widget.MediumImportance
	p.disconnectBtn.Disable()

	buttonBox := container.NewHBox(p.connectBtn, p.disconnectBtn)

	p.container = container.NewVBox(
		title,
		widget.NewLabel(""),
		addrLabel,
		p.addressEntry,
		portLabel,
		p.portEntry,
		widget.NewLabel(""),
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
		p.disconnectBtn.Importance = widget.HighImportance
		p.addressEntry.Disable()
		p.portEntry.Disable()
	} else {
		p.connectBtn.Enable()
		p.connectBtn.Importance = widget.HighImportance
		p.disconnectBtn.Disable()
		p.disconnectBtn.Importance = widget.MediumImportance
		p.addressEntry.Enable()
		p.portEntry.Enable()
	}
	p.connectBtn.Refresh()
	p.disconnectBtn.Refresh()
}

// SetConnecting updates UI to show connecting state.
func (p *ConnectionPanel) SetConnecting(connecting bool) {
	if connecting {
		p.connectBtn.Disable()
		p.disconnectBtn.Enable()
		p.addressEntry.Disable()
		p.portEntry.Disable()
	} else {
		p.connectBtn.Enable()
		p.addressEntry.Enable()
		p.portEntry.Enable()
	}
}

// IsConnected returns the current connection state.
func (p *ConnectionPanel) IsConnected() bool {
	return p.connected
}

// GetAddress returns the current address value.
func (p *ConnectionPanel) GetAddress() string {
	return p.addressEntry.Text
}

// GetPort returns the current port value.
func (p *ConnectionPanel) GetPort() int {
	port, _ := strconv.Atoi(p.portEntry.Text)
	return port
}
