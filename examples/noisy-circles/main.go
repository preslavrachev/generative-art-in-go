package main

import (
	"image/png"
	"log"
	"os"
	"time"

	"github.com/preslavrachev/generative-art-in-go/sketch"
)

func main() {
	seed := time.Now().Unix()

	// Test different fill ratios
	fillRatios := []struct {
		ratio float64
		name  string
	}{
		{1.0, "fully-filled"},
		{0.7, "mostly-filled"},
		{0.5, "half-filled"},
		{0.3, "mostly-outline"},
		{0.0, "all-outline"},
	}

	log.Println("Generating noisy circles with different fill ratios...")

	// Generate both normal and inverted versions
	for _, inverted := range []bool{false, true} {
		suffix := ""
		if inverted {
			suffix = "-inverted"
		}

		for _, fr := range fillRatios {
			params := sketch.NoisyCirclesParams{
				Width:             1200,
				Height:            1200,
				Step:              100,
				Precision:         30,
				Diff:              0.3,
				NoiseAmt:          0.5,
				NoiseScale:        0.8,
				Palette:           sketch.DefaultPalette(),
				Seed:              seed, // Same seed for comparison
				UseFlowField:      true,
				FlowFieldScale:    0.003,
				FlowFieldStrength: 0.99,
				FillRatio:         fr.ratio,
				Inverted:          inverted,
			}

			s := sketch.NewNoisyCircles(params)
			img := s.Generate()

			filename := "noisy-circles-" + fr.name + suffix + ".png"
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}

			if err := png.Encode(f, img); err != nil {
				f.Close()
				log.Fatal(err)
			}
			f.Close()

			log.Printf("✓ Generated %s (FillRatio=%.1f, Inverted=%v)", filename, fr.ratio, inverted)
		}
	}

	log.Println("\nAll variations generated!")
	log.Println("FillRatio explanation:")
	log.Println("  1.0 = All circles filled with gradients")
	log.Println("  0.5 = Random 50/50 mix of filled and outline")
	log.Println("  0.0 = All circles as colored outlines only")
	log.Println("\nInverted versions use white backgrounds with dark outlines")
	log.Println("Try combining with different palettes for varied aesthetics!")
}
