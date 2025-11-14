package ui

import (
	"game/pkg/bump"
)

// TextboxData holds the pure data state for textbox rendering.
type TextboxData struct {
	// Text is the raw text content to display (may contain line breaks)
	Text string

	// ProcessedText is the text after line wrapping and pagination
	ProcessedText string

	// Lines is the total number of lines in the processed text
	Lines int

	// AdvanceState tracks which page of text is currently displayed
	AdvanceState int

	// AdvanceMax is the maximum page index (0-based)
	AdvanceMax int

	// BoxHeight is the calculated height of the textbox
	BoxHeight int

	// Active indicates whether the textbox is currently visible/interacting
	Active bool

	// Dismissed indicates the player explicitly closed the dialog (down on last page)
	// When dismissed, the textbox won't auto-show again until player presses Up
	Dismissed bool

	// AdvanceFlicker controls the blinking of the advance indicator
	AdvanceFlicker bool

	// AdvanceFlickerTimer tracks time for the flicker animation
	AdvanceFlickerTimer float64

	// Indicator determines if a position indicator should be shown
	Indicator bool

	// Area is the trigger area for textbox activation (nil if not used)
	Area func() bump.Rect

	// EntityX, EntityY, EntityW, EntityH store the entity bounds for indicator positioning
	EntityX, EntityY, EntityW, EntityH float64
}
