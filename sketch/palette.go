package sketch

import (
	"image/color"
	"math"
	"math/rand"
)

// AIDEV-NOTE: Procedural color palette generation using color theory

// PaletteType defines different color harmony schemes
type PaletteType string

const (
	PaletteComplementary PaletteType = "complementary" // Opposite colors on wheel
	PaletteAnalogous     PaletteType = "analogous"     // Adjacent colors
	PaletteTriadic       PaletteType = "triadic"       // Evenly spaced (120°)
	PaletteSplitComp     PaletteType = "split-comp"    // Base + two adjacent to complement
	PaletteTetradic      PaletteType = "tetradic"      // Two complementary pairs
	PaletteMonochromatic PaletteType = "monochromatic" // Single hue, varied saturation/lightness
	PaletteWarm          PaletteType = "warm"          // Warm colors (reds, oranges, yellows)
	PaletteCool          PaletteType = "cool"          // Cool colors (blues, greens, purples)
	PalettePastel        PaletteType = "pastel"        // Soft, desaturated colors
	PaletteVibrant       PaletteType = "vibrant"       // High saturation colors
)

// GeneratePalette creates a color palette based on the specified type
func GeneratePalette(paletteType PaletteType, count int) []color.Color {
	baseHue := rand.Float64() * 360 // Random starting hue

	switch paletteType {
	case PaletteComplementary:
		return generateComplementary(baseHue, count)
	case PaletteAnalogous:
		return generateAnalogous(baseHue, count)
	case PaletteTriadic:
		return generateTriadic(baseHue, count)
	case PaletteSplitComp:
		return generateSplitComplementary(baseHue, count)
	case PaletteTetradic:
		return generateTetradic(baseHue, count)
	case PaletteMonochromatic:
		return generateMonochromatic(baseHue, count)
	case PaletteWarm:
		return generateWarm(count)
	case PaletteCool:
		return generateCool(count)
	case PalettePastel:
		return generatePastel(count)
	case PaletteVibrant:
		return generateVibrant(count)
	default:
		return DefaultPalette()
	}
}

func generateComplementary(baseHue float64, count int) []color.Color {
	colors := []color.Color{}
	complement := math.Mod(baseHue+180, 360)

	// Alternate between base and complement with variations
	for i := 0; i < count; i++ {
		var hue float64
		if i%2 == 0 {
			hue = baseHue + rand.Float64()*30 - 15 // ±15° variation
		} else {
			hue = complement + rand.Float64()*30 - 15
		}

		sat := 0.6 + rand.Float64()*0.3  // 60-90% saturation
		light := 0.5 + rand.Float64()*0.3 // 50-80% lightness

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateAnalogous(baseHue float64, count int) []color.Color {
	colors := []color.Color{}
	spread := 60.0 // ±30° from base

	for i := 0; i < count; i++ {
		hue := baseHue + (rand.Float64()*spread - spread/2)
		hue = math.Mod(hue+360, 360)

		sat := 0.5 + rand.Float64()*0.4
		light := 0.45 + rand.Float64()*0.35

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateTriadic(baseHue float64, count int) []color.Color {
	colors := []color.Color{}
	hues := []float64{baseHue, math.Mod(baseHue+120, 360), math.Mod(baseHue+240, 360)}

	for i := 0; i < count; i++ {
		hue := hues[i%3] + rand.Float64()*20 - 10 // ±10° variation

		sat := 0.6 + rand.Float64()*0.35
		light := 0.5 + rand.Float64()*0.3

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateSplitComplementary(baseHue float64, count int) []color.Color {
	colors := []color.Color{}
	complement := math.Mod(baseHue+180, 360)
	split1 := math.Mod(complement-30, 360)
	split2 := math.Mod(complement+30, 360)
	hues := []float64{baseHue, split1, split2}

	for i := 0; i < count; i++ {
		hue := hues[i%3] + rand.Float64()*15 - 7.5

		sat := 0.55 + rand.Float64()*0.35
		light := 0.45 + rand.Float64()*0.4

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateTetradic(baseHue float64, count int) []color.Color {
	colors := []color.Color{}
	hues := []float64{
		baseHue,
		math.Mod(baseHue+90, 360),
		math.Mod(baseHue+180, 360),
		math.Mod(baseHue+270, 360),
	}

	for i := 0; i < count; i++ {
		hue := hues[i%4] + rand.Float64()*15 - 7.5

		sat := 0.5 + rand.Float64()*0.4
		light := 0.45 + rand.Float64()*0.35

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateMonochromatic(baseHue float64, count int) []color.Color {
	colors := []color.Color{}

	for i := 0; i < count; i++ {
		hue := baseHue + rand.Float64()*10 - 5 // Very slight hue variation

		sat := 0.3 + rand.Float64()*0.5           // Varied saturation
		light := 0.3 + float64(i)*0.5/float64(count) // Gradient from dark to light

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateWarm(count int) []color.Color {
	colors := []color.Color{}

	for i := 0; i < count; i++ {
		// Warm hues: red to yellow (0-60°)
		hue := rand.Float64() * 60
		if rand.Float64() > 0.7 {
			// Occasional orange-red (330-360°)
			hue = 330 + rand.Float64()*30
		}

		sat := 0.6 + rand.Float64()*0.35
		light := 0.5 + rand.Float64()*0.3

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateCool(count int) []color.Color {
	colors := []color.Color{}

	for i := 0; i < count; i++ {
		// Cool hues: cyan to violet (180-300°)
		hue := 180 + rand.Float64()*120

		sat := 0.5 + rand.Float64()*0.4
		light := 0.45 + rand.Float64()*0.35

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generatePastel(count int) []color.Color {
	colors := []color.Color{}

	for i := 0; i < count; i++ {
		hue := rand.Float64() * 360

		sat := 0.25 + rand.Float64()*0.35 // Low saturation
		light := 0.7 + rand.Float64()*0.2  // High lightness

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

func generateVibrant(count int) []color.Color {
	colors := []color.Color{}

	for i := 0; i < count; i++ {
		hue := rand.Float64() * 360

		sat := 0.8 + rand.Float64()*0.2  // Very high saturation
		light := 0.45 + rand.Float64()*0.25 // Medium lightness

		colors = append(colors, hslToRGB(hue, sat, light))
	}

	return colors
}

// hslToRGB converts HSL color space to RGB
func hslToRGB(h, s, l float64) color.Color {
	h = math.Mod(h, 360) / 360.0

	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l // Achromatic
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}

		p := 2*l - q

		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
		A: 255,
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
