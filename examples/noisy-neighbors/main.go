package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/fogleman/gg"
	"github.com/preslavrachev/generative-art-in-go/sketch"
	"github.com/preslavrachev/generative-art-in-go/util"
)

func main() {
	seed := time.Now().Unix()

	log.Println("Generating Noisy Neighbors with force simulation...")

	params := NoisyNeighborsParams{
		Width:      1200,
		Height:     1200,
		Count:      500, // Number of neighbors
		BaseRadius: 50,
		Precision:  30,
		Diff:       0.4,
		NoiseAmt:   0.5,
		NoiseScale: 0.8,
		Palette:    sketch.DefaultPalette(),
		Seed:       seed,

		// Flow field settings
		UseFlowField:      true,
		FlowFieldScale:    0.002,
		FlowFieldStrength: 0.7,

		// Force simulation settings
		AttractionForce: 0.5, // Gentle attraction to flow field
		RepulsionForce:  800, // Strong repulsion from neighbors
		RepulsionRadius: 120, // Distance at which neighbors repel
		SimulationSteps: 200, // Iterations to settle

		// Density-based variation
		DensityAffectsSize: true,
		MinRadiusScale:     0.4, // Circles in dense areas shrink to 40% of base size

		// Rendering
		FillRatio: 0.7,
		Inverted:  false,
	}

	s := NewNoisyNeighbors(params)
	img := s.Generate()

	filename := "noisy-neighbors.png"
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}
	f.Close()

	log.Printf("✓ Generated %s", filename)

	// Generate inverted version
	log.Println("Generating inverted version...")
	params.Inverted = true
	params.Seed = seed // Same seed for comparison

	s = NewNoisyNeighbors(params)
	img = s.Generate()

	filename = "noisy-neighbors-inverted.png"
	f, err = os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}
	f.Close()

	log.Printf("✓ Generated %s", filename)

	log.Println("\nNoisy Neighbors features:")
	log.Println("  • Attraction/repulsion force simulation creates organic clustering")
	log.Println("  • Density-based size variation: circles shrink in crowded areas")
	log.Println("  • Flow field guides overall movement and color distribution")
	log.Println("  • Each 'neighbor' finds its space while influenced by the community")
}

// AIDEV-NOTE: Noisy neighbors with force simulation and density-based variations

type Neighbor struct {
	x, y        float64 // Current position
	vx, vy      float64 // Velocity
	radius      float64 // Base radius (varies by density)
	gradientIdx int     // Index into gradients array
	noiseZ      float64
	flowAngle   float64
}

type NoisyNeighborsParams struct {
	Width             int
	Height            int
	Count             int // Number of neighbors
	BaseRadius        float64
	Precision         int
	Diff              float64
	NoiseAmt          float64
	NoiseScale        float64
	Palette           []color.Color
	Seed              int64
	UseFlowField      bool
	FlowFieldScale    float64
	FlowFieldStrength float64
	FillRatio         float64
	Inverted          bool
	// Force simulation params
	AttractionForce float64 // Strength of attraction to flow field
	RepulsionForce  float64 // Strength of repulsion from neighbors
	RepulsionRadius float64 // Distance at which repulsion kicks in
	SimulationSteps int     // Number of force simulation iterations
	// Density-based variation
	DensityAffectsSize bool    // Whether crowded areas have smaller circles
	MinRadiusScale     float64 // Minimum radius multiplier in dense areas (e.g., 0.5)
}

type NoisyNeighborsSketch struct {
	params    NoisyNeighborsParams
	dc        *gg.Context
	neighbors []Neighbor
	gradients []gg.Gradient
	perlin    *perlin.Perlin
}

func NewNoisyNeighbors(params NoisyNeighborsParams) *NoisyNeighborsSketch {
	s := &NoisyNeighborsSketch{
		params: params,
	}

	rand.Seed(params.Seed)
	s.perlin = perlin.NewPerlin(2.0, 2.0, 3, params.Seed)

	return s
}

func (s *NoisyNeighborsSketch) Generate() image.Image {
	s.dc = gg.NewContext(s.params.Width, s.params.Height)

	s.initializeNeighbors()
	s.createGradients()
	s.runForceSimulation()
	s.calculateDensities()
	s.draw()

	return s.dc.Image()
}

func (s *NoisyNeighborsSketch) initializeNeighbors() {
	s.neighbors = make([]Neighbor, s.params.Count)

	// Distribute neighbors randomly across canvas
	for i := 0; i < s.params.Count; i++ {
		s.neighbors[i] = Neighbor{
			x:           rand.Float64() * float64(s.params.Width),
			y:           rand.Float64() * float64(s.params.Height),
			vx:          0,
			vy:          0,
			radius:      s.params.BaseRadius,
			gradientIdx: i,
			noiseZ:      rand.Float64()*400 + 100,
			flowAngle:   0,
		}
	}
}

func (s *NoisyNeighborsSketch) createGradients() {
	s.gradients = make([]gg.Gradient, s.params.Count)

	for i := 0; i < s.params.Count; i++ {
		n := &s.neighbors[i]

		gx := rand.Float64()*(s.params.BaseRadius*0.5) - s.params.BaseRadius*0.25
		gy := rand.Float64()*(s.params.BaseRadius*0.5) - s.params.BaseRadius*0.25

		grad := gg.NewRadialGradient(gx, gy, 0, gx, gy, s.params.BaseRadius*2)

		var color1, color2 color.Color

		if s.params.UseFlowField {
			noiseVal := s.perlin.Noise2D(n.x*s.params.FlowFieldScale, n.y*s.params.FlowFieldScale)
			flowAngle := (noiseVal + 1) * math.Pi
			n.flowAngle = flowAngle

			normalizedAngle := math.Mod(flowAngle, 2*math.Pi) / (2 * math.Pi)
			palettePos := normalizedAngle * float64(len(s.params.Palette))

			idx1 := int(math.Floor(palettePos)) % len(s.params.Palette)
			if idx1 < 0 {
				idx1 += len(s.params.Palette)
			}
			idx2 := (idx1 + 1) % len(s.params.Palette)

			color1 = s.params.Palette[idx1]
			color2 = s.params.Palette[idx2]
		} else {
			color1 = s.params.Palette[rand.Intn(len(s.params.Palette))]
			color2 = s.params.Palette[rand.Intn(len(s.params.Palette))]
		}

		grad.AddColorStop(0, color1)
		grad.AddColorStop(1, color2)

		s.gradients[i] = grad
	}
}

func (s *NoisyNeighborsSketch) runForceSimulation() {
	for step := 0; step < s.params.SimulationSteps; step++ {
		for i := range s.neighbors {
			n := &s.neighbors[i]

			// Reset forces
			fx, fy := 0.0, 0.0

			// Attraction to flow field direction
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
					other := &s.neighbors[j]

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

			// Apply forces with damping
			damping := 0.7
			n.vx = (n.vx + fx) * damping
			n.vy = (n.vy + fy) * damping

			// Update position
			n.x += n.vx
			n.y += n.vy

			// Bounce off walls
			margin := s.params.BaseRadius * 2
			if n.x < margin {
				n.x = margin
				n.vx *= -0.5
			}
			if n.x > float64(s.params.Width)-margin {
				n.x = float64(s.params.Width) - margin
				n.vx *= -0.5
			}
			if n.y < margin {
				n.y = margin
				n.vy *= -0.5
			}
			if n.y > float64(s.params.Height)-margin {
				n.y = float64(s.params.Height) - margin
				n.vy *= -0.5
			}
		}
	}
}

func (s *NoisyNeighborsSketch) calculateDensities() {
	if !s.params.DensityAffectsSize {
		return
	}

	// For each neighbor, count nearby neighbors to determine local density
	for i := range s.neighbors {
		n := &s.neighbors[i]

		neighborCount := 0
		searchRadius := s.params.RepulsionRadius * 1.5

		for j := range s.neighbors {
			if i == j {
				continue
			}
			other := &s.neighbors[j]

			dx := n.x - other.x
			dy := n.y - other.y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < searchRadius {
				neighborCount++
			}
		}

		// Scale radius inversely with density
		maxNeighbors := 8.0
		densityRatio := math.Min(float64(neighborCount)/maxNeighbors, 1.0)
		radiusScale := 1.0 - (densityRatio * (1.0 - s.params.MinRadiusScale))

		n.radius = s.params.BaseRadius * radiusScale
	}
}

func (s *NoisyNeighborsSketch) draw() {
	if s.params.Inverted {
		s.dc.SetRGB(1, 1, 1)
	} else {
		s.dc.SetRGB(0.04, 0.04, 0.04)
	}
	s.dc.Clear()

	for i := range s.neighbors {
		n := &s.neighbors[i]

		s.dc.Push()
		s.dc.Translate(n.x, n.y)

		// Decide whether to fill this circle
		shouldFill := rand.Float64() < s.params.FillRatio

		if shouldFill {
			s.dc.SetFillStyle(s.gradients[n.gradientIdx])
		}

		// Calculate rotation
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

		// Draw main circle
		if shouldFill {
			s.noisyCircle(n.radius, n.noiseZ, true)
		} else {
			s.dc.SetStrokeStyle(s.gradients[n.gradientIdx])
			s.dc.SetLineWidth(2)
			s.noisyCircle(n.radius, n.noiseZ, false)
		}

		// Additional outline circles
		if s.params.Inverted {
			s.dc.SetRGB(0.04, 0.04, 0.04)
		} else {
			s.dc.SetRGB(1, 1, 1)
		}
		s.dc.SetLineWidth(2)
		s.noisyCircle(n.radius, n.noiseZ+s.params.Diff, false)
		s.noisyCircle(n.radius, n.noiseZ+s.params.Diff*2, false)

		s.dc.Pop()
	}
}

func (s *NoisyNeighborsSketch) noisyCircle(r, noiseZ float64, fill bool) {
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
		n = util.MapRange(n, 0, 1, -noiseAmount, noiseAmount)

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
