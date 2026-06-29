package sketch

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
	"github.com/fogleman/gg"
	"github.com/preslavrachev/generative-art-in-go/util"
)

// AIDEV-NOTE: Image-driven brush stroke system with multi-centroid particle spawning

type ImageBrushstrokesParams struct {
	SourceImage     image.Image
	OutputWidth     int
	OutputHeight    int
	ParticleCount   int // Total particles to spawn
	CentroidCount   int // Number of spawn points (10-15 recommended)
	SimulationSteps int // How many frames to simulate
	MaxTrailLength  int // Max stroke segments per particle
	Seed            int64

	// Force parameters
	RadialForce       float64 // Outward push from centroids
	FlowFieldScale    float64 // Perlin noise scale for flow field
	FlowFieldStrength float64 // How much flow field affects movement
	RepulsionForce    float64 // Particle-to-particle repulsion
	RepulsionRadius   float64 // Distance threshold for repulsion
	Damping           float64 // Velocity damping (0.7-0.9)

	// Rendering parameters
	StrokeWidth    float64 // Base width of brush marks
	StrokeAlpha    float64 // Opacity (0.7-0.9 recommended)
	ColorBlendRate float64 // How fast particle color updates from source (0.2-0.4)
	VelocityScale  float64 // Stroke length multiplier based on speed
}

type Particle struct {
	x, y         float64 // Position in output canvas space
	vx, vy       float64 // Velocity
	currentColor color.Color
	centroidIdx  int // Which centroid spawned this
	trail        []TrailSegment
	age          int
}

type TrailSegment struct {
	x, y   float64
	angle  float64
	width  float64
	length float64
	color  color.Color
}

type Centroid struct {
	x, y  float64 // Position in output canvas space
	score float64 // Interest score from source image
}

type ImageBrushstrokesSketch struct {
	params    ImageBrushstrokesParams
	dc        *gg.Context
	particles []*Particle
	centroids []Centroid
	perlin    *perlin.Perlin
}

func NewImageBrushstrokes(params ImageBrushstrokesParams) *ImageBrushstrokesSketch {
	s := &ImageBrushstrokesSketch{
		params: params,
	}

	rand.Seed(params.Seed)
	s.perlin = perlin.NewPerlin(2.0, 2.0, 3, params.Seed)

	return s
}

func (s *ImageBrushstrokesSketch) Generate() image.Image {
	s.dc = gg.NewContext(s.params.OutputWidth, s.params.OutputHeight)

	// Setup
	s.detectCentroids()
	s.spawnParticles()

	// Simulate and paint
	for step := 0; step < s.params.SimulationSteps; step++ {
		s.updateParticles()
	}

	// Render all trails
	s.draw()

	return s.dc.Image()
}

// detectCentroids finds interesting regions in source image
func (s *ImageBrushstrokesSketch) detectCentroids() {
	bounds := s.params.SourceImage.Bounds()
	sourceW := bounds.Dx()
	sourceH := bounds.Dy()

	// Sample grid to find high-interest points
	sampleStep := 20 // Sample every 20 pixels
	candidates := []Centroid{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStep {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStep {
			c := s.params.SourceImage.At(x, y)
			r, g, b, _ := c.RGBA()

			// Convert to [0,1] range
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0

			// Calculate interest score (saturation + brightness)
			maxRGB := math.Max(rf, math.Max(gf, bf))
			minRGB := math.Min(rf, math.Min(gf, bf))
			saturation := 0.0
			if maxRGB > 0 {
				saturation = (maxRGB - minRGB) / maxRGB
			}
			brightness := (rf + gf + bf) / 3.0

			score := saturation*0.6 + brightness*0.4

			// Map to output space
			outX := float64(x-bounds.Min.X) * float64(s.params.OutputWidth) / float64(sourceW)
			outY := float64(y-bounds.Min.Y) * float64(s.params.OutputHeight) / float64(sourceH)

			candidates = append(candidates, Centroid{
				x:     outX,
				y:     outY,
				score: score,
			})
		}
	}

	// Sort by score and take top N
	// Simple bubble sort (good enough for small lists)
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Take top CentroidCount
	count := s.params.CentroidCount
	if count > len(candidates) {
		count = len(candidates)
	}
	s.centroids = candidates[:count]
}

// spawnParticles distributes particles across centroids
func (s *ImageBrushstrokesSketch) spawnParticles() {
	s.particles = make([]*Particle, s.params.ParticleCount)

	for i := 0; i < s.params.ParticleCount; i++ {
		// Assign to random centroid
		centroidIdx := rand.Intn(len(s.centroids))
		centroid := s.centroids[centroidIdx]

		// Spawn near centroid with some jitter
		jitterRadius := 30.0
		angle := rand.Float64() * 2 * math.Pi
		jitter := rand.Float64() * jitterRadius

		x := centroid.x + math.Cos(angle)*jitter
		y := centroid.y + math.Sin(angle)*jitter

		// Sample initial color
		initialColor := s.sampleColor(x, y)

		s.particles[i] = &Particle{
			x:            x,
			y:            y,
			vx:           0,
			vy:           0,
			currentColor: initialColor,
			centroidIdx:  centroidIdx,
			trail:        []TrailSegment{},
			age:          0,
		}
	}
}

// sampleColor maps output canvas coords to source image and samples color
func (s *ImageBrushstrokesSketch) sampleColor(x, y float64) color.Color {
	bounds := s.params.SourceImage.Bounds()

	// Map output space to source space
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

// updateParticles runs one simulation step
func (s *ImageBrushstrokesSketch) updateParticles() {
	for _, p := range s.particles {
		// Forces
		fx, fy := 0.0, 0.0

		// 1. Radial force from birth centroid
		centroid := s.centroids[p.centroidIdx]
		dx := p.x - centroid.x
		dy := p.y - centroid.y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > 1.0 {
			// Normalize and apply radial force
			fx += (dx / dist) * s.params.RadialForce
			fy += (dy / dist) * s.params.RadialForce
		}

		// 2. Flow field influence
		noiseVal := s.perlin.Noise2D(p.x*s.params.FlowFieldScale, p.y*s.params.FlowFieldScale)
		flowAngle := (noiseVal + 1) * math.Pi
		fx += math.Cos(flowAngle) * s.params.FlowFieldStrength
		fy += math.Sin(flowAngle) * s.params.FlowFieldStrength

		// 3. Repulsion from nearby particles
		for _, other := range s.particles {
			if p == other {
				continue
			}

			dx := p.x - other.x
			dy := p.y - other.y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < s.params.RepulsionRadius && dist > 0.1 {
				force := s.params.RepulsionForce / (dist * dist)
				fx += (dx / dist) * force
				fy += (dy / dist) * force
			}
		}

		// Apply forces with damping
		p.vx = (p.vx + fx) * s.params.Damping
		p.vy = (p.vy + fy) * s.params.Damping

		// Update position
		oldX, oldY := p.x, p.y
		p.x += p.vx
		p.y += p.vy

		// Bounce off edges
		margin := 20.0
		if p.x < margin {
			p.x = margin
			p.vx *= -0.5
		}
		if p.x > float64(s.params.OutputWidth)-margin {
			p.x = float64(s.params.OutputWidth) - margin
			p.vx *= -0.5
		}
		if p.y < margin {
			p.y = margin
			p.vy *= -0.5
		}
		if p.y > float64(s.params.OutputHeight)-margin {
			p.y = float64(s.params.OutputHeight) - margin
			p.vy *= -0.5
		}

		// Sample new color and blend
		sampledColor := s.sampleColor(p.x, p.y)
		p.currentColor = blendColors(p.currentColor, sampledColor, s.params.ColorBlendRate)

		// Add trail segment
		speed := math.Sqrt(p.vx*p.vx + p.vy*p.vy)
		angle := math.Atan2(p.vy, p.vx)
		length := speed * s.params.VelocityScale

		if length > 0.5 { // Only draw if moving
			segment := TrailSegment{
				x:      (oldX + p.x) / 2, // Midpoint
				y:      (oldY + p.y) / 2,
				angle:  angle,
				width:  s.params.StrokeWidth,
				length: length,
				color:  p.currentColor,
			}

			p.trail = append(p.trail, segment)

			// Limit trail length
			if len(p.trail) > s.params.MaxTrailLength {
				p.trail = p.trail[1:]
			}
		}

		p.age++
	}
}

// blendColors mixes two colors
func blendColors(c1, c2 color.Color, rate float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8(util.MapRange(rate, 0, 1, float64(r1>>8), float64(r2>>8)))
	g := uint8(util.MapRange(rate, 0, 1, float64(g1>>8), float64(g2>>8)))
	b := uint8(util.MapRange(rate, 0, 1, float64(b1>>8), float64(b2>>8)))
	a := uint8(util.MapRange(rate, 0, 1, float64(a1>>8), float64(a2>>8)))

	return color.RGBA{r, g, b, a}
}

// draw renders all particle trails
func (s *ImageBrushstrokesSketch) draw() {
	// Background
	s.dc.SetRGB(0.98, 0.98, 0.98)
	s.dc.Clear()

	// Render all trails
	for _, p := range s.particles {
		for _, segment := range p.trail {
			s.drawStroke(segment)
		}
	}
}

// drawStroke renders a single rectangular brush mark
func (s *ImageBrushstrokesSketch) drawStroke(seg TrailSegment) {
	s.dc.Push()

	// Set color with alpha
	r, g, b, _ := seg.color.RGBA()
	s.dc.SetRGBA(
		float64(r)/65535.0,
		float64(g)/65535.0,
		float64(b)/65535.0,
		s.params.StrokeAlpha,
	)

	s.dc.Translate(seg.x, seg.y)
	s.dc.Rotate(seg.angle)

	// Draw rectangle centered at origin
	s.dc.DrawRectangle(-seg.length/2, -seg.width/2, seg.length, seg.width)
	s.dc.Fill()

	s.dc.Pop()
}
