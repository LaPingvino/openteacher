// Package backgroundimage provides functionality for generating background images
// used in OpenTeacher's UI and promotional materials.
//
// This is a Go port of the Python BackgroundImageGenerator module from:
// legacy/modules/org/openteacher/profileRunners/backgroundImageGenerator/backgroundImageGenerator.py
package backgroundimage

import (
	"context"
	"fmt"
	"github.com/LaPingvino/openteacher/internal/core"
	"image"
	"image/color"
	"image/draw"
	"math"
)

// BackgroundImageGeneratorModule handles the generation of background images
type BackgroundImageGeneratorModule struct {
	*core.BaseModule
	manager *core.Manager

	// Module-specific properties
	hue       int
	logo      string
	lineColor color.RGBA
}

// NewBackgroundImageGeneratorModule creates a new BackgroundImageGeneratorModule instance
func NewBackgroundImageGeneratorModule() *BackgroundImageGeneratorModule {
	base := core.NewBaseModule("backgroundImageGenerator", "background-image-generator")
	base.SetRequires("metadata", "ui")

	return &BackgroundImageGeneratorModule{
		BaseModule: base,
	}
}

// Enable activates the module
// This is the Go equivalent of the Python enable method
func (big *BackgroundImageGeneratorModule) Enable(ctx context.Context) error {
	if err := big.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// TODO: Get modules interface from module manager
	// TODO: Get metadata module to retrieve:
	//   - mainColorHue for big.hue
	//   - iconPath for big.logo

	// For now, set some default values
	big.hue = 200 // Default blue hue
	big.logo = "" // TODO: Get from metadata
	big.lineColor = big.colorFromHSV(big.hue, 28, 186)

	fmt.Println("BackgroundImageGenerator module enabled")
	return nil
}

// Disable deactivates the module
// This is the Go equivalent of the Python disable method
func (big *BackgroundImageGeneratorModule) Disable(ctx context.Context) error {
	if err := big.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up resources
	big.hue = 0
	big.logo = ""
	big.lineColor = color.RGBA{}

	fmt.Println("BackgroundImageGenerator module disabled")
	return nil
}

// SetManager sets the module manager
func (big *BackgroundImageGeneratorModule) SetManager(manager *core.Manager) {
	big.manager = manager
}

// Generate creates the body background image, which includes:
//   - a rounded lighter area, on which the content is shown
//   - the logo
//   - the application name
//
// This is the Go equivalent of the Python generate method
func (big *BackgroundImageGeneratorModule) Generate() (image.Image, error) {
	if !big.IsActive() {
		return nil, fmt.Errorf("module is not active")
	}

	// Set image dimensions and layout values
	width := 1000
	height := 5000
	sideMargin := 27
	topMargin := 64

	xRadius := 9
	yRadius := int(float64(xRadius) * 0.7)
	topLineY := 94

	logoTopX := 6
	logoSize := 107

	textXStart := 124
	textYBaseline := 58

	// Determine colors based on hue
	textColor := big.colorFromHSV(big.hue, 119, 47)
	gradientTopColor := big.colorFromHSV(big.hue, 7, 253)
	gradientBottomColor := big.colorFromHSV(big.hue, 12, 243)

	// Create image with transparent background
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with transparent
	transparent := color.RGBA{0, 0, 0, 0}
	draw.Draw(img, img.Bounds(), &image.Uniform{transparent}, image.ZP, draw.Src)

	// Draw rounded rectangle with gradient
	err := big.drawRoundedRectangleWithGradient(
		img,
		sideMargin, topMargin,
		width-sideMargin*2, height,
		xRadius, yRadius,
		gradientTopColor, gradientBottomColor,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to draw rounded rectangle: %w", err)
	}

	// Draw top line
	big.drawLine(img, sideMargin, topLineY, width-sideMargin, topLineY, big.lineColor)

	// Draw logo
	err = big.drawLogo(img, logoTopX, 0, logoSize)
	if err != nil {
		return nil, fmt.Errorf("failed to draw logo: %w", err)
	}

	// Draw text
	err = big.drawApplicationText(img, textXStart, textYBaseline, textColor)
	if err != nil {
		return nil, fmt.Errorf("failed to draw application text: %w", err)
	}

	return img, nil
}

// colorFromHSV converts HSV values to RGBA color
// This is the Go equivalent of Qt's QColor.fromHsv
func (big *BackgroundImageGeneratorModule) colorFromHSV(h, s, v int) color.RGBA {
	// Convert HSV to RGB
	// H: 0-359, S: 0-255, V: 0-255
	hf := float64(h%360) / 60.0
	sf := float64(s) / 255.0
	vf := float64(v) / 255.0

	c := vf * sf
	x := c * (1 - math.Abs(math.Mod(hf, 2)-1))
	m := vf - c

	var r, g, b float64

	switch int(hf) {
	case 0:
		r, g, b = c, x, 0
	case 1:
		r, g, b = x, c, 0
	case 2:
		r, g, b = 0, c, x
	case 3:
		r, g, b = 0, x, c
	case 4:
		r, g, b = x, 0, c
	case 5:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// drawRoundedRectangleWithGradient draws a rounded rectangle with gradient fill
// This is the Go equivalent of Qt's painter operations for rounded rectangle and gradient
func (big *BackgroundImageGeneratorModule) drawRoundedRectangleWithGradient(
	img *image.RGBA,
	x, y, width, height, xRadius, yRadius int,
	topColor, bottomColor color.RGBA,
) error {
	// TODO: Implement proper rounded rectangle drawing with gradient
	// This is a simplified version - would need proper rounded corners and gradient

	// Create a simple vertical gradient for now
	for py := y; py < y+height; py++ {
		// Calculate gradient interpolation factor (0.0 to 1.0)
		t := float64(py-y) / float64(height)

		// Interpolate between top and bottom colors
		r := uint8(float64(topColor.R)*(1-t) + float64(bottomColor.R)*t)
		g := uint8(float64(topColor.G)*(1-t) + float64(bottomColor.G)*t)
		b := uint8(float64(topColor.B)*(1-t) + float64(bottomColor.B)*t)
		a := uint8(float64(topColor.A)*(1-t) + float64(bottomColor.A)*t)

		gradientColor := color.RGBA{r, g, b, a}

		// Draw horizontal line with this color
		for px := x; px < x+width; px++ {
			img.Set(px, py, gradientColor)
		}
	}

	// Draw border with lineColor
	// Top and bottom lines
	for px := x; px < x+width; px++ {
		img.Set(px, y, big.lineColor)
		img.Set(px, y+height-1, big.lineColor)
	}
	// Left and right lines
	for py := y; py < y+height; py++ {
		img.Set(x, py, big.lineColor)
		img.Set(x+width-1, py, big.lineColor)
	}

	return nil
}

// drawLine draws a line between two points
// This is the Go equivalent of Qt's painter.drawLine
func (big *BackgroundImageGeneratorModule) drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.RGBA) {
	// Simple horizontal line implementation (most common case for this module)
	if y1 == y2 {
		// Horizontal line
		start, end := x1, x2
		if start > end {
			start, end = end, start
		}
		for x := start; x <= end; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
				img.Set(x, y1, col)
			}
		}
		return
	}

	// Simple vertical line implementation
	if x1 == x2 {
		start, end := y1, y2
		if start > end {
			start, end = end, start
		}
		for y := start; y <= end; y++ {
			if x1 >= 0 && x1 < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x1, y, col)
			}
		}
		return
	}

	// TODO: Implement Bresenham's line algorithm for diagonal lines
	// For now, just draw the endpoints
	if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
		img.Set(x1, y1, col)
	}
	if x2 >= 0 && x2 < img.Bounds().Dx() && y2 >= 0 && y2 < img.Bounds().Dy() {
		img.Set(x2, y2, col)
	}
}

// drawLogo draws the OpenTeacher logo at the specified position and size
// This is the Go equivalent of Qt's painter.drawImage with logo scaling
func (big *BackgroundImageGeneratorModule) drawLogo(img *image.RGBA, x, y, size int) error {
	// TODO: Load and scale logo image
	// This would need to:
	// 1. Load the logo image from big.logo path
	// 2. Scale it to the specified height (size)
	// 3. Draw it at the specified position

	fmt.Printf("Logo drawing not yet implemented - would draw logo at (%d, %d) with size %d\n", x, y, size)

	// For now, draw a placeholder rectangle where the logo would be
	logoRect := image.Rect(x, y, x+size, y+size)
	logoColor := color.RGBA{100, 100, 100, 128} // Semi-transparent gray
	draw.Draw(img, logoRect, &image.Uniform{logoColor}, image.ZP, draw.Over)

	return nil
}

// drawApplicationText draws the "OpenTeacher" text with custom styling
// This is the Go equivalent of Qt's text drawing with different font sizes
func (big *BackgroundImageGeneratorModule) drawApplicationText(img *image.RGBA, startX, baselineY int, textColor color.RGBA) error {
	// TODO: Implement text rendering
	// This would need to render:
	// - "O" in large font (Verdana, 37pt, weight 55)
	// - "PEN" in smaller font (29pt)
	// - "T" in large font
	// - "EACHER" in smaller font
	// - Letter spacing: 95% of normal
	// - Semi-transparent text (alpha: 200)

	fmt.Printf("Text rendering not yet implemented - would draw 'OpenTeacher' text at (%d, %d)\n", startX, baselineY)

	// For now, draw placeholder rectangles where text would be
	// Simulate the different text parts with rectangles of different sizes

	// "O" - large
	oRect := image.Rect(startX, baselineY-30, startX+25, baselineY)
	draw.Draw(img, oRect, &image.Uniform{textColor}, image.ZP, draw.Over)

	// "PEN" - smaller
	penRect := image.Rect(startX+25, baselineY-22, startX+70, baselineY)
	draw.Draw(img, penRect, &image.Uniform{textColor}, image.ZP, draw.Over)

	// "T" - large
	tRect := image.Rect(startX+70, baselineY-30, startX+90, baselineY)
	draw.Draw(img, tRect, &image.Uniform{textColor}, image.ZP, draw.Over)

	// "EACHER" - smaller
	eacherRect := image.Rect(startX+90, baselineY-22, startX+170, baselineY)
	draw.Draw(img, eacherRect, &image.Uniform{textColor}, image.ZP, draw.Over)

	return nil
}

// Init creates and returns a new BackgroundImageGeneratorModule instance
// This is the Go equivalent of the Python init function
func Init() core.Module {
	return NewBackgroundImageGeneratorModule()
}
