package client

// Machine represents an installed virtual machine in vbox.
type Machine struct {
	id string
	vb *VirtualBox
}

// NewMachine returns a pointer to a Machine value
func NewMachine(vb *VirtualBox, id string) *Machine {
	return &Machine{vb: vb, id: id}
}
