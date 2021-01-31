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

type CircularMask struct {
	source image.Image
	center image.Point
	radius int
}

func (c *CircularMask) ColorModel() color.Model {
	return c.source.ColorModel()
}

func (c *CircularMask) Bounds() image.Rectangle {
	return image.Rect(c.center.X-c.radius, c.center.Y-c.radius, c.center.X+c.radius, c.center.Y+c.radius)
}

func (c *CircularMask) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.center.X), float64(y-c.center.Y), float64(c.radius)
	if xx*xx+yy*yy < rr*rr {
		return c.source.At(x, y)
	}
	return color.Alpha{0}
}

func main() {
	img, err := loadImage("source.jpg")
	if err != nil {
		log.Panic(err)
	}

	saveOutput(&CircularMask{
		source: img,
		center: image.Point{
			X: 1000,
			Y: 2000,
		},
		radius: 100,
	}, "out.png")
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
