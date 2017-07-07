package ilda

import (
	"bytes"
	"image"
	"image/png"
	"math"

	"github.com/ungerik/go-cairo"
)

const (
	// 3D Coordinates with Indexed Color
	Format3DIndexColor = Format(0)
	// 2D Coordinates with Indexed Color
	Format2DIndexColor = Format(1)
	// Color Palette for Indexed Color Frames
	FormatPaletteIndex = Format(2)
	// 3D Coordinates with True Color
	Format3DTrueColor = Format(4)
	// 2D Coordinates with True Color
	Format2DTrueColor = Format(5)
)

type Format uint8

func (f Format) String() string {
	switch f {
	case Format3DIndexColor:
		return "3D Coordinates with Indexed Color"
	case Format2DIndexColor:
		return "2D Coordinates with Indexed Color"
	case FormatPaletteIndex:
		return "Color Palette for Indexed Color Frames"
	case Format3DTrueColor:
		return "3D Coordinates with True Color"
	case Format2DTrueColor:
		return "2D Coordinates with True Color"
	default:
		return "Unknown or invalid format"
	}
}

// A struct able to contain the data for the format containing the most
// information, Format3DTrueColor.
type Record struct {
	X, Y, Z int16
	R, G, B uint8
	Status  uint8
}

func (rec Record) Last() bool {
	return rec.Status&(1<<7) != 0
}

func (rec Record) Blank() bool {
	return rec.Status&(1<<6) != 0
}

type Frame struct {
	Format      Format
	FrameName   string
	CompanyName string
	Records     []Record
}

func (frame *Frame) Image(w, h int) image.Image {
	surf := cairo.NewSurface(cairo.FORMAT_RGB24, w, h)
	surf.Scale(float64(w)/float64(math.MaxUint16), float64(h)/float64(math.MaxUint16))
	surf.Translate(float64(math.MaxInt16), float64(math.MaxInt16))
	surf.Scale(1, -1)
	surf.SetLineWidth(float64(math.MaxUint16) / float64(w))

	for i, rec := range frame.Records[1:] {
		if rec.Last() {
			break
		} else if rec.Blank() {
			continue
		}
		prevRec := frame.Records[i]
		surf.SetSourceRGB(float64(rec.R)/255, float64(rec.G)/255, float64(rec.B)/255)
		surf.MoveTo(float64(prevRec.X), float64(prevRec.Y))
		surf.LineTo(float64(rec.X), float64(rec.Y))
		surf.Stroke()
	}

	// Surface has a GetImage() method, but it seems to be broken. :(
	// It would be even better if Surface implemented the Image interface.
	pngBytes, _ := surf.WriteToPNGStream()
	img, _ := png.Decode(bytes.NewReader(pngBytes))
	return img
}
