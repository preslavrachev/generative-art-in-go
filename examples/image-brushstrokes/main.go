package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/fogleman/gg"
	"github.com/preslavrachev/generative-art-in-go/util"
)

func main() {
	sourceImg, err := loadImage("source.jpg")
	if err != nil {
		log.Fatalf("failed to load source image: %v", err)
	}

	bounds := sourceImg.Bounds()
	log.Printf("Source image loaded: %dx%d", bounds.Dx(), bounds.Dy())

	// Sample a test color to verify image is loaded correctly
	testColor := sourceImg.At(bounds.Dx()/2, bounds.Dy()/2)
	r, g, b, _ := testColor.RGBA()
	log.Printf("Center pixel color: R=%d G=%d B=%d", r>>8, g>>8, b>>8)

	// Sample multiple points
	for i := 0; i < 5; i++ {
		x := rand.Intn(bounds.Dx())
		y := rand.Intn(bounds.Dy())
		c := sourceImg.At(x, y)
		r, g, b, _ := c.RGBA()
		log.Printf("Sample at (%d,%d): R=%d G=%d B=%d", x, y, r>>8, g>>8, b>>8)
	}

	log.Println("Generating image-based noisy neighbors...")

	seed := time.Now().Unix()
	rand.Seed(seed)

	params := Params{
		SourceImage:  sourceImg,
		OutputWidth:  2400,
		OutputHeight: 2400,
		Count:        1000,
		BaseRadius:   100,
		Precision:    50,
		Diff:         0.4,
		NoiseAmt:     0.5,
		NoiseScale:   0.8,
		Seed:         seed,

		UseFlowField:      true,
		FlowFieldScale:    0.002,
		FlowFieldStrength: 0.7,

		AttractionForce: 0.5,
		RepulsionForce:  800,
		RepulsionRadius: 120,
		SimulationSteps: 200,

		DensityAffectsSize: true,
		MinRadiusScale:     0.4,

		FillRatio: 0.7,

		TrailAlpha:    0.12,
		TrailInterval: 0,
	}

	s := NewSketch(params)
	img := s.Generate()

	filename := "image-noisy-neighbors.png"
	if err := saveImage(img, filename); err != nil {
		log.Fatalf("failed to save: %v", err)
	}

	log.Printf("✓ Generated %s", filename)
}

type Params struct {
	SourceImage        image.Image
	OutputWidth        int
	OutputHeight       int
	Count              int
	BaseRadius         float64
	Precision          int
	Diff               float64
	NoiseAmt           float64
	NoiseScale         float64
	Seed               int64
	UseFlowField       bool
	FlowFieldScale     float64
	FlowFieldStrength  float64
	FillRatio          float64
	AttractionForce    float64
	RepulsionForce     float64
	RepulsionRadius    float64
	SimulationSteps    int
	DensityAffectsSize bool
	MinRadiusScale     float64
	TrailAlpha         float64 // Alpha for intermediate trail rendering
	TrailInterval      int     // Draw trails every N steps
}

type Neighbor struct {
	x, y      float64
	vx, vy    float64
	radius    float64
	gradient  gg.Gradient
	noiseZ    float64
	flowAngle float64
}

type Sketch struct {
	params    Params
	dc        *gg.Context
	neighbors []*Neighbor
	perlin    *perlin.Perlin
}

func NewSketch(params Params) *Sketch {
	return &Sketch{
		params: params,
		perlin: perlin.NewPerlin(2.0, 2.0, 3, params.Seed),
	}
}

func (s *Sketch) Generate() image.Image {
	s.dc = gg.NewContext(s.params.OutputWidth, s.params.OutputHeight)

	// White background
	s.dc.SetRGB(1, 1, 1)
	s.dc.Clear()

	s.initializeNeighbors()
	s.runForceSimulationWithTrails()
	s.createGradients()
	s.calculateDensities()
	s.drawFinal()

	return s.dc.Image()
}

func (s *Sketch) initializeNeighbors() {
	s.neighbors = make([]*Neighbor, s.params.Count)

	for i := 0; i < s.params.Count; i++ {
		s.neighbors[i] = &Neighbor{
			x:         rand.Float64() * float64(s.params.OutputWidth),
			y:         rand.Float64() * float64(s.params.OutputHeight),
			vx:        0,
			vy:        0,
			radius:    s.params.BaseRadius,
			noiseZ:    rand.Float64()*400 + 100,
			flowAngle: 0,
		}
	}
}

func (s *Sketch) runForceSimulationWithTrails() {
	// Create temporary gradients for trail rendering
	s.createGradients()

	for step := 0; step < s.params.SimulationSteps; step++ {
		// Simulate one step
		for i := range s.neighbors {
			n := s.neighbors[i]

			fx, fy := 0.0, 0.0

			// Flow field attraction
			if s.params.UseFlowField && s.params.AttractionForce > 0 {
				noiseVal := s.perlin.Noise2D(n.x*s.params.FlowFieldScale, n.y*s.params.FlowFieldScale)
				angle := (noiseVal + 1) * math.Pi
				fx += math.Cos(angle) * s.params.AttractionForce
				fy += math.Sin(angle) * s.params.AttractionForce
			}

			// Repulsion from neighbors
			if s.params.RepulsionForce > 0 {
				for j := range s.neighbors {
					if i == j {
						continue
					}
					other := s.neighbors[j]

					dx := n.x - other.x
					dy := n.y - other.y
					dist := math.Sqrt(dx*dx + dy*dy)

					if dist < s.params.RepulsionRadius && dist > 0.1 {
						force := s.params.RepulsionForce / (dist * dist)
						fx += (dx / dist) * force
						fy += (dy / dist) * force
					}
				}
			}

			damping := 0.7
			n.vx = (n.vx + fx) * damping
			n.vy = (n.vy + fy) * damping

			n.x += n.vx
			n.y += n.vy

			// Bounce off walls
			margin := s.params.BaseRadius * 2
			if n.x < margin {
				n.x = margin
				n.vx *= -0.5
			}
			if n.x > float64(s.params.OutputWidth)-margin {
				n.x = float64(s.params.OutputWidth) - margin
				n.vx *= -0.5
			}
			if n.y < margin {
				n.y = margin
				n.vy *= -0.5
			}
			if n.y > float64(s.params.OutputHeight)-margin {
				n.y = float64(s.params.OutputHeight) - margin
				n.vy *= -0.5
			}
		}

		// Paint trails at intervals
		if step%s.params.TrailInterval == 0 {
			s.drawTrailFrame()
		}
	}
}

func (s *Sketch) createGradients() {
	for _, n := range s.neighbors {
		gx := rand.Float64()*(s.params.BaseRadius*0.5) - s.params.BaseRadius*0.25
		gy := rand.Float64()*(s.params.BaseRadius*0.5) - s.params.BaseRadius*0.25

		grad := gg.NewRadialGradient(gx, gy, 0, gx, gy, s.params.BaseRadius*2)

		// Sample colors from source image at neighbor position
		color1 := s.sampleColor(n.x, n.y)

		// Sample nearby for second color
		offsetX := n.x + rand.Float64()*s.params.BaseRadius*2 - s.params.BaseRadius
		offsetY := n.y + rand.Float64()*s.params.BaseRadius*2 - s.params.BaseRadius
		color2 := s.sampleColor(offsetX, offsetY)

		if s.params.UseFlowField {
			noiseVal := s.perlin.Noise2D(n.x*s.params.FlowFieldScale, n.y*s.params.FlowFieldScale)
			n.flowAngle = (noiseVal + 1) * math.Pi
		}

		// Convert to RGBA to ensure proper color handling
		r1, g1, b1, a1 := color1.RGBA()
		r2, g2, b2, a2 := color2.RGBA()

		c1 := color.RGBA{uint8(r1 >> 8), uint8(g1 >> 8), uint8(b1 >> 8), uint8(a1 >> 8)}
		c2 := color.RGBA{uint8(r2 >> 8), uint8(g2 >> 8), uint8(b2 >> 8), uint8(a2 >> 8)}

		grad.AddColorStop(0, c1)
		grad.AddColorStop(1, c2)

		n.gradient = grad
	}
}

func (s *Sketch) sampleColor(x, y float64) color.Color {
	bounds := s.params.SourceImage.Bounds()

	// Map output coords to source coords
	sx := x * float64(bounds.Dx()) / float64(s.params.OutputWidth)
	sy := y * float64(bounds.Dy()) / float64(s.params.OutputHeight)

	px := int(sx) + bounds.Min.X
	py := int(sy) + bounds.Min.Y

	// Clamp
	if px < bounds.Min.X {
		px = bounds.Min.X
	}
	if px >= bounds.Max.X {
		px = bounds.Max.X - 1
	}
	if py < bounds.Min.Y {
		py = bounds.Min.Y
	}
	if py >= bounds.Max.Y {
		py = bounds.Max.Y - 1
	}

	return s.params.SourceImage.At(px, py)
}

func (s *Sketch) calculateDensities() {
	if !s.params.DensityAffectsSize {
		return
	}

	for i := range s.neighbors {
		n := s.neighbors[i]

		neighborCount := 0
		searchRadius := s.params.RepulsionRadius * 1.5

		for j := range s.neighbors {
			if i == j {
				continue
			}
			other := s.neighbors[j]

			dx := n.x - other.x
			dy := n.y - other.y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < searchRadius {
				neighborCount++
			}
		}

		maxNeighbors := 8.0
		densityRatio := math.Min(float64(neighborCount)/maxNeighbors, 1.0)
		radiusScale := 1.0 - (densityRatio * (1.0 - s.params.MinRadiusScale))

		n.radius = s.params.BaseRadius * radiusScale
	}
}

func (s *Sketch) drawTrailFrame() {
	for _, n := range s.neighbors {
		s.dc.Push()
		s.dc.Translate(n.x, n.y)

		// Set gradient with trail alpha
		r1, g1, b1, _ := n.gradient.ColorAt(0, 0).RGBA()
		s.dc.SetRGBA(
			float64(r1)/65535.0,
			float64(g1)/65535.0,
			float64(b1)/65535.0,
			s.params.TrailAlpha,
		)

		var rotationAngle float64
		if s.params.UseFlowField {
			flowAngle := n.flowAngle
			noiseAngle := s.perlin.Noise2D(n.y+n.radius, n.x-n.y) * 2 * math.Pi
			rotationAngle = flowAngle*s.params.FlowFieldStrength +
				noiseAngle*(1-s.params.FlowFieldStrength)
		} else {
			noiseAngle := s.perlin.Noise2D(n.y+n.radius, n.x-n.y)
			rotationAngle = noiseAngle * 2 * math.Pi
		}

		s.dc.Rotate(rotationAngle)

		// Just draw filled circle for trails
		s.noisyCircle(n.radius, n.noiseZ, true)

		s.dc.Pop()
	}
}

func (s *Sketch) drawFinal() {
	for _, n := range s.neighbors {
		s.dc.Push()
		s.dc.Translate(n.x, n.y)

		shouldFill := rand.Float64() < s.params.FillRatio

		if shouldFill {
			s.dc.SetFillStyle(n.gradient)
		}

		var rotationAngle float64
		if s.params.UseFlowField {
			flowAngle := n.flowAngle
			noiseAngle := s.perlin.Noise2D(n.y+n.radius, n.x-n.y) * 2 * math.Pi
			rotationAngle = flowAngle*s.params.FlowFieldStrength +
				noiseAngle*(1-s.params.FlowFieldStrength)
		} else {
			noiseAngle := s.perlin.Noise2D(n.y+n.radius, n.x-n.y)
			rotationAngle = noiseAngle * 2 * math.Pi
		}

		s.dc.Rotate(rotationAngle)

		if shouldFill {
			s.noisyCircle(n.radius, n.noiseZ, true)
		} else {
			s.dc.SetStrokeStyle(n.gradient)
			s.dc.SetLineWidth(2)
			s.noisyCircle(n.radius, n.noiseZ, false)
		}

		// Dark outlines
		s.dc.SetRGB(0.04, 0.04, 0.04)
		s.dc.SetLineWidth(2)
		s.noisyCircle(n.radius, n.noiseZ+s.params.Diff, false)
		s.noisyCircle(n.radius, n.noiseZ+s.params.Diff*2, false)

		s.dc.Pop()
	}
}

func (s *Sketch) noisyCircle(r, noiseZ float64, fill bool) {
	angleStep := 2 * math.Pi / float64(s.params.Precision)
	noiseAmount := r * s.params.NoiseAmt
	noiseScale := s.params.NoiseScale / r

	points := make([][2]float64, 0, s.params.Precision+4)

	for i := -1; i <= s.params.Precision+1; i++ {
		angle := angleStep * float64(i)
		x := math.Cos(angle) * r
		y := math.Sin(angle) * r

		n := s.perlin.Noise3D(noiseScale*x, noiseScale*y, noiseZ)
		n = util.MapRange(n, 0, 1, -noiseAmount, noiseAmount)

		points = append(points, [2]float64{x + n, y + n})
	}

	s.dc.MoveTo(points[1][0], points[1][1])

	for i := 1; i < len(points)-2; i++ {
		p0 := points[i-1]
		p1 := points[i]
		p2 := points[i+1]
		p3 := points[i+2]

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

func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	return img, nil
}

func saveImage(img image.Image, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}
