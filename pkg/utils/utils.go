// Package text provides text rendering utilities for the game.
// These functions handle text measurement and rendering with support for
// different font faces.
package utils

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// TextSize calculates the width and height of the given text when rendered with the provided face.
func TextSize(txt string, face text.Face) (float64, float64) {
	size := faceSize(face)
	w, h := text.Measure(txt, face, size+1)
	return w - 1, h
}

// DrawText renders text to the given image using the provided face and draw options.
// Renders directly using text.Draw() with no position manipulation, matching the
// debug overlay text rendering pattern.
//
// Returns the width and height of the rendered text.
func DrawText(img *ebiten.Image, txt string, face text.Face, imgOp *ebiten.DrawImageOptions) (float64, float64) {
	op := &text.DrawOptions{}
	if imgOp != nil {
		op.DrawImageOptions = *imgOp
	}

	op.LineSpacing = faceSize(face) + 1
	text.Draw(img, txt, face, op)

	return TextSize(txt, face)
}

// faceSize attempts to extract a size value from the provided text.Face.
// It understands *text.GoTextFace and any type exposing Size() float64.
// Falls back to 16 if the concrete type doesn't expose size information.
func faceSize(f text.Face) float64 {
	switch v := f.(type) {
	case *text.GoTextFace:
		return v.Size
	case interface{ Size() float64 }:
		return v.Size()
	default:
		return 16
	}
}
