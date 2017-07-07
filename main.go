package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
	"sync"

	"github.com/polyfloyd/ilda/ilda"
)

const IMAGE_OUT_WIDTH = 512
const IMAGE_OUT_HEIGHT = 512

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <ilda file>\n", os.Args[0])
		return
	}
	inputFilename := os.Args[1]
	fd, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer fd.Close()

	frames, err := ilda.Decode(fd)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	gifImage := gif.GIF{
		Image: make([]*image.Paletted, len(frames)),
		Delay: make([]int, len(frames)),
	}
	var wg sync.WaitGroup
	wg.Add(len(frames))
	for i, frame := range frames {
		go func(i int, frame ilda.Frame) {
			palettedImage := image.NewPaletted(image.Rect(0, 0, IMAGE_OUT_WIDTH, IMAGE_OUT_HEIGHT), palette.Plan9)
			renderedFrame := frame.Image(IMAGE_OUT_WIDTH, IMAGE_OUT_HEIGHT)
			draw.Draw(palettedImage, palettedImage.Rect, renderedFrame, image.Pt(0, 0), draw.Over)
			gifImage.Image[i] = palettedImage
			wg.Done()
		}(i, frame)
	}
	wg.Wait()

	out, err := os.Create(inputFilename + ".gif")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if err := gif.EncodeAll(out, &gifImage); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
}
