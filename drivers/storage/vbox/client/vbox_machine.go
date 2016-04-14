package client

// Machine represents an installed virtual machine in vbox.
type Machine struct {
	mobref string
	id     string
	name   string
	vb     *VirtualBox
}

// NewMachine returns a pointer to a Machine value
func NewMachine(vb *VirtualBox, id string) *Machine {
	return &Machine{vb: vb, id: id}
}

// GetID returns the ID last populated for this machine
func (m *Machine) GetID() string {
	return m.id
}

// GetName returns the Name last populated for this machine
func (m *Machine) GetName() string {
	return m.name
}
