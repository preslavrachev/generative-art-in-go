package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

type GrayscaleFilter struct {
	image.Image
}

func (f *GrayscaleFilter) At(x, y int) color.Color {
	r, g, b, a := f.Image.At(x, y).RGBA()
	grey := uint16(float64(r)*0.21 + float64(g)*0.72 + float64(b)*0.07)

	return color.RGBA64{
		R: grey,
		G: grey,
		B: grey,
		A: uint16(a),
	}
}

func main() {
	img, err := loadImage("source.jpg")
	if err != nil {
		log.Panic(err)
	}

	saveOutput(&GrayscaleFilter{img}, "out.png")
}

func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("source image could not be loaded: %w", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("source image format could not be decoded: %w", err)
	}

	return img, nil
}

func saveOutput(img image.Image, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode to `PNG` with `DefaultCompression` level
	// then save to file
	err = png.Encode(f, img)
	if err != nil {
		return err
	}

	return nil
}
