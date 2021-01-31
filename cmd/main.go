package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/preslavrachev/generative-art-in-go/sketch"
)

var (
	sourceImgName   = "source.jpg"
	outputImgName   = "out.png"
	totalCycleCount = 5000
)

func main() {

	img, err := loadImage(sourceImgName)
	/*
		use this for completely random results
		img, err := loadRandomUnsplashImage(2000, 2000)
	*/
	if err != nil {
		log.Panicln(err)
	}

	destWidth := 2000
	s := sketch.NewSketch(img, sketch.UserParams{
		StrokeRatio:              0.75,
		DestWidth:                destWidth,
		DestHeight:               2000,
		InitialAlpha:             0.1,
		StrokeReduction:          0.002,
		AlphaIncrease:            0.06,
		StrokeInversionThreshold: 0.05,
		StrokeJitter:             int(0.1 * float64(destWidth)),
		MinEdgeCount:             3,
		MaxEdgeCount:             4,
	})

	rand.Seed(time.Now().Unix())

	for i := 0; i < totalCycleCount; i++ {
		s.Update()
	}

	saveOutput(s.Output(), outputImgName)
}

func loadRandomUnsplashImage(width, height int) (image.Image, error) {
	url := fmt.Sprintf("https://source.unsplash.com/random/%dx%d", width, height)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	return img, err
}

func loadImage(src string) (image.Image, error) {
	file, _ := os.Open(sourceImgName)
	defer file.Close()
	img, _, err := image.Decode(file)
	return img, err
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
