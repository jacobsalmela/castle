package visual

// VisualEffects contains visual feedback state for entities.
// Used for flash effects when taking damage, status effects, etc.
type VisualEffects struct {
	// FlashTimer counts down from FlashDuration to 0.
	// While > 0, entity should render with flash effect.
	FlashTimer float64

	// FlashDuration is the total duration of the flash effect in seconds.
	// Set this during initialization to control flash length.
	FlashDuration float64

	// FlashColor is the RGB color of the flash effect.
	// Default is white [1, 1, 1], but can be customized per entity.
	// Common colors:
	//   {1, 0, 0} = Red
	//   {0, 1, 0} = Green
	//   {0, 0, 1} = Blue
	//   {1, 1, 1} = White
	//   {1, 1, 0} = Yellow
	FlashColor [3]float32
}
