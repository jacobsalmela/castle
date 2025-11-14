package config

import "github.com/hajimehoshi/ebiten/v2"

// Input holds input configuration including key bindings.
type Input struct {
	// Key bindings map logical actions to physical keys
	// Each action can have multiple keys bound to it
	Right  []string `yaml:"right" line_comment:"Move right keys"`
	Left   []string `yaml:"left" line_comment:"Move left keys"`
	Up     []string `yaml:"up" line_comment:"Move up / climb keys"`
	Down   []string `yaml:"down" line_comment:"Move down / drop keys"`
	Jump   []string `yaml:"jump" line_comment:"Jump keys"`
	Action []string `yaml:"action" line_comment:"Attack / interact keys"`
	Guard  []string `yaml:"guard" line_comment:"Block / guard keys"`
	Heal   []string `yaml:"heal" line_comment:"Heal keys"`
	Dash   []string `yaml:"dash" line_comment:"Dash keys"`
}

// NewDefaultInput returns default key bindings.
func NewDefaultInput() Input {
	return Input{
		Right:  []string{"ArrowRight", "D"},
		Left:   []string{"ArrowLeft", "A"},
		Up:     []string{"ArrowUp", "W"},
		Down:   []string{"ArrowDown", "S"},
		Jump:   []string{"Z", "N"},
		Action: []string{"X", "M"},
		Guard:  []string{"C", "B"},
		Heal:   []string{"V", "ControlLeft", "Control"},
		Dash:   []string{"Space"},
	}
}

// KeyNameToEbitenKey converts a key name string to an ebiten.Key.
// Returns -1 if the key name is not recognized.
func KeyNameToEbitenKey(name string) ebiten.Key {
	switch name {
	// Arrow keys
	case "ArrowRight":
		return ebiten.KeyArrowRight
	case "ArrowLeft":
		return ebiten.KeyArrowLeft
	case "ArrowUp":
		return ebiten.KeyArrowUp
	case "ArrowDown":
		return ebiten.KeyArrowDown

	// Letter keys
	case "A":
		return ebiten.KeyA
	case "B":
		return ebiten.KeyB
	case "C":
		return ebiten.KeyC
	case "D":
		return ebiten.KeyD
	case "E":
		return ebiten.KeyE
	case "F":
		return ebiten.KeyF
	case "G":
		return ebiten.KeyG
	case "H":
		return ebiten.KeyH
	case "I":
		return ebiten.KeyI
	case "J":
		return ebiten.KeyJ
	case "K":
		return ebiten.KeyK
	case "L":
		return ebiten.KeyL
	case "M":
		return ebiten.KeyM
	case "N":
		return ebiten.KeyN
	case "O":
		return ebiten.KeyO
	case "P":
		return ebiten.KeyP
	case "Q":
		return ebiten.KeyQ
	case "R":
		return ebiten.KeyR
	case "S":
		return ebiten.KeyS
	case "T":
		return ebiten.KeyT
	case "U":
		return ebiten.KeyU
	case "V":
		return ebiten.KeyV
	case "W":
		return ebiten.KeyW
	case "X":
		return ebiten.KeyX
	case "Y":
		return ebiten.KeyY
	case "Z":
		return ebiten.KeyZ

	// Number keys
	case "0", "Digit0":
		return ebiten.KeyDigit0
	case "1", "Digit1":
		return ebiten.KeyDigit1
	case "2", "Digit2":
		return ebiten.KeyDigit2
	case "3", "Digit3":
		return ebiten.KeyDigit3
	case "4", "Digit4":
		return ebiten.KeyDigit4
	case "5", "Digit5":
		return ebiten.KeyDigit5
	case "6", "Digit6":
		return ebiten.KeyDigit6
	case "7", "Digit7":
		return ebiten.KeyDigit7
	case "8", "Digit8":
		return ebiten.KeyDigit8
	case "9", "Digit9":
		return ebiten.KeyDigit9

	// Special keys
	case "Space":
		return ebiten.KeySpace
	case "Enter":
		return ebiten.KeyEnter
	case "Escape":
		return ebiten.KeyEscape
	case "Tab":
		return ebiten.KeyTab
	case "Backspace":
		return ebiten.KeyBackspace

	// Modifier keys
	case "Shift", "ShiftLeft":
		return ebiten.KeyShiftLeft
	case "ShiftRight":
		return ebiten.KeyShiftRight
	case "Control", "ControlLeft":
		return ebiten.KeyControlLeft
	case "ControlRight":
		return ebiten.KeyControlRight
	case "Alt", "AltLeft":
		return ebiten.KeyAltLeft
	case "AltRight":
		return ebiten.KeyAltRight

	default:
		return -1
	}
}

// ToEbitenKeys converts the Input config to ebiten.Key arrays for each logical key.
// Returns a [9][]ebiten.Key array indexed by InputKey constants.
func (i *Input) ToEbitenKeys() [9][]ebiten.Key {
	var bindings [9][]ebiten.Key

	// Helper to convert string slice to ebiten.Key slice
	convert := func(names []string) []ebiten.Key {
		var keys []ebiten.Key
		for _, name := range names {
			if key := KeyNameToEbitenKey(name); key >= 0 {
				keys = append(keys, key)
			}
		}
		return keys
	}

	bindings[0] = convert(i.Right)  // InputKeyRight
	bindings[1] = convert(i.Left)   // InputKeyLeft
	bindings[2] = convert(i.Up)     // InputKeyUp
	bindings[3] = convert(i.Down)   // InputKeyDown
	bindings[4] = convert(i.Jump)   // InputKeyJump
	bindings[5] = convert(i.Action) // InputKeyAction
	bindings[6] = convert(i.Guard)  // InputKeyGuard
	bindings[7] = convert(i.Heal)   // InputKeyHeal
	bindings[8] = convert(i.Dash)   // InputKeyDash

	return bindings
}
