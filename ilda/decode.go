package ilda

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
)

type rawHeader struct {
	Ilda        [4]byte
	_           [3]byte
	Format      uint8
	FrameName   [8]byte
	CompanyName [8]byte
	NumRecords  uint16
	FrameNumber uint16
	TotalFrames uint16
	Projector   uint8
	_           uint8
}

func Decode(r io.Reader) ([]Frame, error) {
	frames := []Frame{}
	for {
		var header rawHeader
		if err := binary.Read(r, binary.BigEndian, &header); err != nil {
			return nil, err
		}

		// The frame should start with the string "ILDA".
		if string(header.Ilda[:]) != "ILDA" {
			return nil, fmt.Errorf("File header does not start with ILDA")
		}
		if header.NumRecords == 0 {
			break // EOF
		}

		frame := Frame{
			Format:      Format(header.Format),
			FrameName:   string(bytes.TrimRight(header.FrameName[:], "\x00")),
			CompanyName: string(bytes.TrimRight(header.CompanyName[:], "\x00")),
		}

		var err error
		switch frame.Format {
		case Format3DIndexColor:
			frame.Records, err = decodeRecords3DIndexColor(r, int(header.NumRecords), DefaultPalette)
		default:
			err = fmt.Errorf("Unsupported format: %v", frame.Format)
		}
		if err != nil {
			return nil, err
		}

		frames = append(frames, frame)
	}
	return frames, nil
}

func decodeRecords3DIndexColor(r io.Reader, numRecords int, palette color.Palette) ([]Record, error) {
	records := make([]Record, numRecords)
	for i := range records {
		var rawRec struct {
			X, Y, Z    int16
			Status     uint8
			ColorIndex uint8
		}
		if err := binary.Read(r, binary.BigEndian, &rawRec); err != nil {
			return nil, err
		}
		rec := &records[i]
		rec.X, rec.Y, rec.Z = rawRec.X, rawRec.Y, rawRec.Z
		rec.Status = rawRec.Status

		colorIndex := int(rawRec.ColorIndex)
		if colorIndex >= len(palette) {
			return nil, fmt.Errorf("Color index (%d) is greater than the pallette size (%d)", colorIndex, len(palette))
		}
		r, g, b, _ := palette[colorIndex].RGBA()
		rec.R, rec.G, rec.B = uint8(r/256), uint8(g/256), uint8(b/256)
	}
	return records, nil
}
