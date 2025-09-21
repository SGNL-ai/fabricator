package pipeline

import (
	"math"
	"math/rand"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// ClusteringStrategy defines how FK values are selected from target entities
type ClusteringStrategy interface {
	// SelectTargetIndex returns the target row index for a given source row
	// sourceIndex: which source row we're processing (0, 1, 2, ...)
	// targetCount: total number of target rows available
	// cardinality: relationship cardinality (OneToOne, ManyToOne, etc.)
	SelectTargetIndex(sourceIndex, targetCount int, cardinality string) int
}

// RoundRobinStrategy implements simple round-robin distribution
type RoundRobinStrategy struct{}

func (s *RoundRobinStrategy) SelectTargetIndex(sourceIndex, targetCount int, cardinality string) int {
	if targetCount == 0 {
		return 0
	}
	return sourceIndex % targetCount
}

// PowerLawStrategy implements power law clustering for realistic distribution
type PowerLawStrategy struct {
	rng *rand.Rand
}

func NewPowerLawStrategy(seed int64) *PowerLawStrategy {
	return &PowerLawStrategy{
		rng: rand.New(rand.NewSource(seed)),
	}
}

func (s *PowerLawStrategy) SelectTargetIndex(sourceIndex, targetCount int, cardinality string) int {
	if targetCount == 0 {
		return 0
	}

	switch cardinality {
	case model.OneToOne:
		// 1:1 relationships should use unique assignment
		return sourceIndex % targetCount

	case model.ManyToOne, model.OneToMany:
		// Use power law distribution for realistic clustering
		return s.powerLawIndex(targetCount)

	default:
		// Fallback to round-robin for unknown cardinalities
		return sourceIndex % targetCount
	}
}

// powerLawIndex generates an index following power law distribution
// This creates realistic clustering where some targets are much more popular
func (s *PowerLawStrategy) powerLawIndex(targetCount int) int {
	if targetCount <= 1 {
		return 0
	}

	// Generate power law distributed value
	// Lower alpha = more concentration (more clustering)
	// Higher alpha = more uniform (less clustering)
	alpha := 1.5 // Moderate clustering

	// Use inverse transform sampling for power law distribution
	// Formula: x = ((1-u)^(-1/(alpha-1)) - 1) where u is uniform random
	u := s.rng.Float64()
	if u == 0 {
		u = 0.001 // Avoid division by zero
	}

	// Transform to power law
	powerValue := math.Pow(1-u, -1/(alpha-1)) - 1

	// Scale to target count and clamp to valid range
	index := int(powerValue * float64(targetCount-1))
	if index >= targetCount {
		index = targetCount - 1
	}
	if index < 0 {
		index = 0
	}

	return index
}

// WeightedRandomStrategy implements weighted random selection with configurable weights
type WeightedRandomStrategy struct {
	rng     *rand.Rand
	weights []float64
}

func NewWeightedRandomStrategy(seed int64, weights []float64) *WeightedRandomStrategy {
	return &WeightedRandomStrategy{
		rng:     rand.New(rand.NewSource(seed)),
		weights: weights,
	}
}

func (s *WeightedRandomStrategy) SelectTargetIndex(sourceIndex, targetCount int, cardinality string) int {
	if targetCount == 0 {
		return 0
	}

	if cardinality == model.OneToOne {
		return sourceIndex % targetCount
	}

	// Use weighted random selection
	if len(s.weights) != targetCount {
		// Fallback to round-robin if weights don't match target count
		return sourceIndex % targetCount
	}

	// Select index based on weights
	totalWeight := 0.0
	for _, weight := range s.weights {
		totalWeight += weight
	}

	if totalWeight <= 0 {
		return sourceIndex % targetCount
	}

	r := s.rng.Float64() * totalWeight
	cumulative := 0.0
	for i, weight := range s.weights {
		cumulative += weight
		if r <= cumulative {
			return i
		}
	}

	return targetCount - 1 // Fallback
}
