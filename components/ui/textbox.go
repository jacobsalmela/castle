package ui

// Textbox component holds data for rendering and managing a UI textbox.
type Textbox struct {
	State  TextboxState
	Layout TextboxLayout
}

// TextboxState captures the runtime textbox state in data-only form.
type TextboxState struct {
	Text           string
	Active         bool
	Indicator      bool
	AdvanceState   int
	AdvanceMax     int
	FlickerOn      bool
	FlickerTimer   float64
	VisibleLines   int
	BoxPixelHeight int
}

// TextboxLayout stores derived layout metrics for rendering alignment.
type TextboxLayout struct {
	BoxWidth int
}
