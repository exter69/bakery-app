package domain

// PanelState represents the internal state of an order or reservation side panel.
// This models the panel state as a pure Go struct for property testing.
type PanelState struct {
	SelectedDay string
	StartTime   string
	EndTime     string
	Items       []OrderItem
	IsOpen      bool
}

// InitialPanelState returns the initial empty state of a side panel.
func InitialPanelState() PanelState {
	return PanelState{}
}

// ResetPanelState resets a panel state back to its initial empty state,
// regardless of the current state contents.
func ResetPanelState(state PanelState) PanelState {
	return InitialPanelState()
}
