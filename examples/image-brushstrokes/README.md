# Image Brushstrokes

Generates brush stroke paintings from source images using multi-centroid particle simulation.

## How It Works

1. **Centroid Detection**: Analyzes source image to find 10-15 "interesting" regions (high saturation/brightness)
2. **Particle Spawning**: Distributes particles across these centroids
3. **Force Simulation**: Particles pushed outward by radial force + flow field + repulsion
4. **Trail Painting**: Each particle paints rectangular brush marks along its path
5. **Color Sampling**: Colors continuously sampled from source image with temporal blending

## Usage

### 1. Prepare Source Image

Place your source image as `source.jpg` (or `source.png`) in this directory.

**Best results with:**
- 1200x1200+ pixels (square preferred)
- High contrast subject on simpler background
- Vibrant, saturated colors
- Central/focal composition

**Examples:**
- Portrait with colorful clothing
- Flowers/still life with dark background
- Abstract art with clear color zones

### 2. Run

```bash
cd examples/image-brushstrokes
go run main.go
```

Output: `image-brushstrokes.png` (4800x4800 pixels by default)

### 3. Tweak Parameters

Edit `main.go` to adjust:

**Canvas Size:**
```go
OutputWidth:   4800,  // Bigger = more detail, slower
OutputHeight:  4800,
```

**Particle Count:**
```go
ParticleCount: 3000,  // More = denser strokes
CentroidCount: 12,    // 10-15 spawn points recommended
```

**Stroke Appearance:**
```go
StrokeWidth:    8,    // Brush thickness
StrokeAlpha:    0.75, // Transparency (0.6-0.9)
ColorBlendRate: 0.3,  // Color persistence (0.2-0.4)
VelocityScale:  2.5,  // Length from speed (2-4)
```

**Movement:**
```go
RadialForce:       0.8,  // Outward push (0.5-1.5)
FlowFieldStrength: 0.5,  // Chaos level (0.3-0.8)
RepulsionRadius:   80,   // Particle spacing (60-120)
```

**Simulation:**
```go
SimulationSteps: 300,  // More = longer trails
MaxTrailLength:  200,  // Segments per particle
```

## Tips

- Start with a small output size (1200x1200) for fast iteration
- Scale up to 4800x4800+ for final prints
- Lower `StrokeAlpha` (0.6) for softer, more blended look
- Higher `RepulsionRadius` (100+) for more distinct strokes
- More `CentroidCount` (15+) for fuller coverage
- Adjust `ColorBlendRate`: lower = more color coherence per stroke

## Output Size Guide

- **1200x1200**: Quick tests (~5 seconds)
- **2400x2400**: Good quality (~15 seconds)
- **4800x4800**: Print quality (~60 seconds)
- **8000x8000**: Gallery quality (~3 minutes)

Processing time scales with `OutputWidth * OutputHeight * ParticleCount * SimulationSteps`.
