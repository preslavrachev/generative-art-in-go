package main

import (
	"image/png"
	"log"
	"os"
	"time"

	"github.com/preslavrachev/generative-art-in-go/sketch"
)

func main() {
	params := sketch.NoisyCirclesParams{
		Width:      1200,
		Height:     1200,
		Step:       100,
		Precision:  30,
		Diff:       0.3,
		NoiseAmt:   0.5,
		NoiseScale: 0.8,
		Palette:    sketch.DefaultPalette(),
		Seed:       time.Now().Unix(),
	}

	s := sketch.NewNoisyCircles(params)
	img := s.Generate()

	f, err := os.Create("noisy-circles.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}

	log.Println("Generated noisy-circles.png")
}
