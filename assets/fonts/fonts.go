package fonts

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	CommonFontFace       text.Face
	RobotoDebugFontFace  text.Face
	RobotoMediumFontFace text.Face
	RobotoBoldFontFace   text.Face
	BitmapDebugFont      *BitmapFont

	commonFontBytes = fonts.MPlus1pRegular_ttf

	//go:embed "m5x7.ttf"
	m5x7File []byte
	//go:embed "m6x11.ttf"
	m6x11File []byte
	//go:embed "nano.ttf"
	nanoFile []byte
	//go:embed "m6x11.ttf"
	debugFile []byte
)

// Init initializes font faces with default DPR of 1.0.
// This is called during asset initialization before config is available.
// Call Ensure(dpr) after config loads to rebuild faces with correct DPR.
func Init() {
	if err := buildFaces(1.0); err != nil {
		log.Fatal(err)
	}
}

var (
	M5x7Font         *text.GoTextFace
	M6x11Font        *text.GoTextFace
	NanoFont         *text.GoTextFace
	FiveByFiveSource *text.GoTextFaceSource
	FiveByFiveFont   *text.GoTextFace
)

func loadFont(data []byte, size float64) *text.GoTextFace {
	face, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

	return &text.GoTextFace{Source: face, Size: size}
}

func buildFaces(dpr float64) error {
	var err error
	CommonFontFace, err = loadFontFace(commonFontBytes, 32*dpr)
	if err != nil {
		return err
	}
	// // Load the bitmap debug font (e.g., 8×8 px, ASCII starting at space)
	// BitmapDebugFont, err = LoadBitmapFont(debugFontPNG, 8, 8, ' ')
	// if err != nil {
	// 	log.Fatal(err)
	// }
	M5x7Font = loadFont(m5x7File, 16)
	M6x11Font = loadFont(m6x11File, 16)
	NanoFont = loadFont(nanoFile, 6*dpr)
	FiveByFiveFont = loadFont(debugFile, 16*dpr) // Reduced from 16 to 8 for better textbox fit
	return nil
}

func loadFontFace(fontData []byte, size float64) (text.Face, error) {
	// optionally, read from a local file
	// ttfFont, err := os.ReadFile(string(fontData))
	// if err != nil {
	// 	return nil, nil
	// }

	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	golangFontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: size,
		DPI:  72,
	})

	// wrap the golang font face into a text/v2.Face
	return text.NewGoXFace(golangFontFace), nil
}

// loadPixelFont creates a font.Face that is exactly pixelSize pixels tall at the given DPI.
func loadPixelFont(fontData []byte, pixelSize, dpi float64) (text.Face, error) {
	// Convert target pixel height to point size:
	ptSize := pixelSize * 72.0 / dpi
	ttfFont, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TTF font: %v", err)
	}
	golangFontFace, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    ptSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	return text.NewGoXFace(golangFontFace), nil
}

func loadTrueTypeFont(fontData []byte, size float64) (text.Face, error) {
	// Parse the TTF font
	otfFont, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TTF font: %v", err)
	}

	// Create a font.Face from the TTF font
	golangFontFace, err := opentype.NewFace(otfFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull, // forces crisp glyph edges
	})
	if err != nil {
		return nil, err
	}
	// wrap the golang font face into a text/v2.Face
	return text.NewGoXFace(golangFontFace), nil
}

func loadOpenTypeFont(fontData []byte, size float64) (text.Face, error) {
	// Parse the TTF font
	otfFont, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TTF font: %v", err)
	}

	// Create a font.Face from the TTF font
	golangFontFace, err := opentype.NewFace(otfFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull, // forces crisp glyph edges
	})
	if err != nil {
		return nil, err
	}
	// wrap the golang font face into a text/v2.Face
	return text.NewGoXFace(golangFontFace), nil
}

// BitmapFont is a simple fixed-width sprite font.
type BitmapFont struct {
	Img        *ebiten.Image
	CharWidth  int
	CharHeight int
	StartRune  rune
	Cols       int
}

// LoadBitmapFont creates a BitmapFont from embedded PNG data.
func LoadBitmapFont(pngData []byte, charWidth, charHeight int, startRune rune) (*BitmapFont, error) {
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode bitmap font PNG: %v", err)
	}
	eimg := ebiten.NewImageFromImage(img)
	bounds := img.Bounds()
	cols := bounds.Dx() / charWidth
	return &BitmapFont{Img: eimg, CharWidth: charWidth, CharHeight: charHeight, StartRune: startRune, Cols: cols}, nil
}

// DrawText draws the string text at (x, y).
func (f *BitmapFont) DrawText(screen *ebiten.Image, textStr string, x, y int) {
	opts := &ebiten.DrawImageOptions{}
	for i, c := range textStr {
		idx := int(c - f.StartRune)
		sx := (idx % f.Cols) * f.CharWidth
		sy := (idx / f.Cols) * f.CharHeight
		subImg := f.Img.SubImage(image.Rect(sx, sy, sx+f.CharWidth, sy+f.CharHeight)).(*ebiten.Image)
		opts.GeoM.Reset()
		opts.GeoM.Translate(float64(x+i*f.CharWidth), float64(y))
		screen.DrawImage(subImg, opts)
	}
}

var (
	mu        sync.Mutex
	cachedDPR = 1.0
	cache     = map[key]text.Face{}
)

type key struct {
	logicalPx float64
	dpr       float64
}

func Ensure(dpr float64) {
	mu.Lock()
	defer mu.Unlock()
	if dpr == cachedDPR {
		return // still valid
	}
	log.Printf("cached: %.2f, dpr: %.2f", cachedDPR, dpr)
	cachedDPR = dpr

	// Close old faces if the concrete type exposes Close().
	for _, f := range cache {
		if c, ok := f.(interface{ Close() error }); ok {
			_ = c.Close() // ignore error; we're about to drop the face anyway
		}
	}
	cache = map[key]text.Face{}

	// Rebuild the faces you need.
	if err := buildFaces(dpr); err != nil {
		log.Fatalf("fonts: rebuild failed: %v", err)
	}
}

// DeviceFace returns a cached face rasterized to `pixelHeight` device pixels at the given DPR.
// pixelHeight is the desired glyph height in device pixels (not logical pixels).
func DeviceFace(pixelHeight int, dpr float64) text.Face {
	mu.Lock()
	defer mu.Unlock()
	k := key{logicalPx: float64(pixelHeight), dpr: dpr}
	if f, ok := cache[k]; ok {
		return f
	}

	// To make the underlying opentype face produce glyphs sized to `pixelHeight` device pixels,
	// pass pixelSize scaled by DPR and set DPI to 72 * DPR so the point-size calculation yields
	// a point size equal to pixelHeight.
	pixelSize := float64(pixelHeight) * dpr
	dpi := dpr * 72.0
	f, err := loadPixelFont(debugFile, pixelSize, dpi)
	if err != nil {
		log.Printf("fonts: DeviceFace failed: %v", err)
		return NanoFont
	}
	cache[k] = f
	return f
}
