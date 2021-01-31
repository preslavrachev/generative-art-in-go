package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

func main() {
	img, err := loadImage("source.jpg")
	if err != nil {
		log.Panic(err)
	}

	saveOutput(resizeImage(img, 500, 500), "out.png")
}

func resizeImage(in image.Image, newWidth int, newHeight int) image.Image {
	originalWidth, originalHeight := in.Bounds().Dx(), in.Bounds().Dy()
	scalingRatioX := float64(originalWidth) / float64(newWidth)
	scalingRatioY := float64(originalHeight) / float64(newHeight)

	out := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for x := 0; x < newWidth; x++ {
		for y := 0; y < newHeight; y++ {
			projectedX := int(float64(x) * scalingRatioX)
			projectedY := int(float64(y) * scalingRatioY)
			out.Set(x, y, in.At(projectedX, projectedY))
		}
	}

	return out
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
