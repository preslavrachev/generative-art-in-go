package sketch

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
	"github.com/fogleman/gg"
)

// AIDEV-NOTE: Params for generative noisy circles with radial gradients

type NoisyCirclesParams struct {
	Width      int
	Height     int
	Step       float64
	Precision  int
	Diff       float64
	NoiseAmt   float64
	NoiseScale float64
	Palette    []color.Color
	Seed       int64
}

type NoisyCirclesSketch struct {
	params    NoisyCirclesParams
	dc        *gg.Context
	gradients []gg.Gradient
	noiseZs   []float64
	perlin    *perlin.Perlin
	m         float64
	radius    float64
}

func NewNoisyCircles(params NoisyCirclesParams) *NoisyCirclesSketch {
	s := &NoisyCirclesSketch{
		params: params,
	}

	rand.Seed(params.Seed)
	s.perlin = perlin.NewPerlin(2.0, 2.0, 3, params.Seed)

	return s
}

func (s *NoisyCirclesSketch) Generate() image.Image {
	s.dc = gg.NewContext(s.params.Width, s.params.Height)

	s.calculateDimensions()
	s.createGradients()
	s.draw()

	return s.dc.Image()
}

func (s *NoisyCirclesSketch) calculateDimensions() {
	minDim := float64(s.params.Width)
	if s.params.Height < s.params.Width {
		minDim = float64(s.params.Height)
	}

	s.m = math.Floor(minDim * 0.9)
	s.m = math.Floor(s.m/s.params.Step) * s.params.Step
	s.radius = s.params.Step * 0.4
}

func (s *NoisyCirclesSketch) createGradients() {
	s.gradients = []gg.Gradient{}
	s.noiseZs = []float64{}

	for x := 0.0; x < s.m; x += s.params.Step {
		for y := 0.0; y < s.m; y += s.params.Step {
			gx := rand.Float64()*(s.params.Step*0.5) - s.params.Step*0.25
			gy := rand.Float64()*(s.params.Step*0.5) - s.params.Step*0.25

			grad := gg.NewRadialGradient(gx, gy, 0, gx, gy, s.params.Step)

			color1 := s.params.Palette[rand.Intn(len(s.params.Palette))]
			color2 := s.params.Palette[rand.Intn(len(s.params.Palette))]

			grad.AddColorStop(0, color1)
			grad.AddColorStop(1, color2)

			s.gradients = append(s.gradients, grad)
			s.noiseZs = append(s.noiseZs, rand.Float64()*400+100)
		}
	}
}

func (s *NoisyCirclesSketch) draw() {
	s.dc.SetRGB(0.04, 0.04, 0.04)
	s.dc.Clear()

	s.dc.Push()
	translateX := (float64(s.params.Width) - s.m + s.params.Step) / 2
	translateY := (float64(s.params.Height) - s.m + s.params.Step) / 2
	s.dc.Translate(translateX, translateY)

	i := 0
	for x := 0.0; x < s.m; x += s.params.Step {
		for y := 0.0; y < s.m; y += s.params.Step {
			s.dc.Push()

			s.dc.SetFillStyle(s.gradients[i])
			s.dc.Translate(x, y)

			noiseAngle := s.perlin.Noise2D(y+s.params.Step, x-y)
			s.dc.Rotate(noiseAngle * 2 * math.Pi)

			s.noisyCircle(s.radius, s.noiseZs[i], true)

			s.dc.SetRGB(1, 1, 1)
			s.dc.SetLineWidth(2)
			s.noisyCircle(s.radius, s.noiseZs[i]+s.params.Diff, false)
			s.noisyCircle(s.radius, s.noiseZs[i]+s.params.Diff*2, false)

			s.dc.Pop()
			i++
		}
	}

	s.dc.Pop()
}

func (s *NoisyCirclesSketch) noisyCircle(r, noiseZ float64, fill bool) {
	angleStep := 2 * math.Pi / float64(s.params.Precision)
	noiseAmount := r * s.params.NoiseAmt
	noiseScale := s.params.NoiseScale / r

	// Calculate all points first for Catmull-Rom curve
	points := make([][2]float64, 0, s.params.Precision+4)

	for i := -1; i <= s.params.Precision+1; i++ {
		angle := angleStep * float64(i)
		x := math.Cos(angle) * r
		y := math.Sin(angle) * r

		n := s.perlin.Noise3D(noiseScale*x, noiseScale*y, noiseZ)
		n = mapRange(n, 0, 1, -noiseAmount, noiseAmount)

		points = append(points, [2]float64{x + n, y + n})
	}

	// Draw using cubic curves (approximates curveVertex behavior)
	s.dc.MoveTo(points[1][0], points[1][1])

	for i := 1; i < len(points)-2; i++ {
		// Simple Catmull-Rom to Bezier conversion
		p0 := points[i-1]
		p1 := points[i]
		p2 := points[i+1]
		p3 := points[i+2]

		// Control points for cubic Bezier
		cp1x := p1[0] + (p2[0]-p0[0])/6.0
		cp1y := p1[1] + (p2[1]-p0[1])/6.0
		cp2x := p2[0] - (p3[0]-p1[0])/6.0
		cp2y := p2[1] - (p3[1]-p1[1])/6.0

		s.dc.CubicTo(cp1x, cp1y, cp2x, cp2y, p2[0], p2[1])
	}

	s.dc.ClosePath()

	if fill {
		s.dc.FillPreserve()
	}
	s.dc.Stroke()
}

func mapRange(value, inMin, inMax, outMin, outMax float64) float64 {
	return (value-inMin)/(inMax-inMin)*(outMax-outMin) + outMin
}

// DefaultPalette returns the color palette from the Processing sketch
func DefaultPalette() []color.Color {
	return []color.Color{
		hexToColor("#f398c3"),
		hexToColor("#cf3895"),
		hexToColor("#a0d28d"),
		hexToColor("#06b4b0"),
		hexToColor("#fed000"),
		hexToColor("#FF8552"),
	}
}

func hexToColor(hex string) color.Color {
	var r, g, b uint8
	if len(hex) == 7 && hex[0] == '#' {
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	}
	return color.RGBA{r, g, b, 255}
}
