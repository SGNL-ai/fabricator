package model

import (
	"errors"
	"fmt"
	"math"
	"os"
)

// Relationship represents a relationship between two entities and their attributes
type Relationship struct {
	id             string
	name           string
	sourceEntity   EntityInterface
	targetEntity   EntityInterface
	sourceAttr     AttributeInterface
	targetAttr     AttributeInterface
	sourceAttrName string // Store attribute names for setup
	targetAttrName string
	cardinality    string
}

// Cardinality constants
const (
	OneToOne  = "1:1"
	OneToMany = "1:N"
	ManyToOne = "N:1"
)

// newRelationship creates a new relationship between entities
// Not exported as only Graph should create relationships
func newRelationship(id, name string, sourceEntity EntityInterface, targetEntity EntityInterface,
	sourceAttributeName string, targetAttributeName string) (RelationshipInterface, error) {

	// 1. Create relationship with initial data
	relationship := &Relationship{
		id:             id,
		name:           name,
		sourceEntity:   sourceEntity,
		targetEntity:   targetEntity,
		sourceAttrName: sourceAttributeName,
		targetAttrName: targetAttributeName,
	}

	// 2. Validate basic properties
	if err := relationship.validate(); err != nil {
		return nil, err
	}

	// 3. Set up attributes
	if err := relationship.setupAttributes(); err != nil {
		return nil, err
	}

	// 4. Determine cardinality
	relationship.determineCardinality()

	return relationship, nil
}

// validate ensures all basic relationship properties are valid
func (r *Relationship) validate() error {
	// Validate basic parameters
	if r.id == "" {
		return errors.New("relationship ID cannot be empty")
	}

	if r.name == "" {
		return errors.New("relationship name cannot be empty")
	}

	// Validate entities
	if r.sourceEntity == nil {
		return errors.New("source entity cannot be nil")
	}

	if r.targetEntity == nil {
		return errors.New("target entity cannot be nil")
	}

	return nil
}

// setupAttributes finds and validates the attributes in the relationship
func (r *Relationship) setupAttributes() error {
	// Look up source attribute
	sourceAttr, exists := r.sourceEntity.GetAttribute(r.sourceAttrName)
	if !exists {
		return fmt.Errorf("source attribute '%s' not found in entity '%s'",
			r.sourceAttrName, r.sourceEntity.GetName())
	}
	r.sourceAttr = sourceAttr

	// Look up target attribute
	targetAttr, exists := r.targetEntity.GetAttribute(r.targetAttrName)
	if !exists {
		return fmt.Errorf("target attribute '%s' not found in entity '%s'",
			r.targetAttrName, r.targetEntity.GetName())
	}
	r.targetAttr = targetAttr

	// Ensure at least one side of the relationship has a unique attribute
	if !r.sourceAttr.IsUnique() && !r.targetAttr.IsUnique() {
		return errors.New("at least one attribute in a relationship must be unique")
	}

	return nil
}

// GetID returns relationship's ID
func (r *Relationship) GetID() string {
	return r.id
}

// GetName returns relationship's name
func (r *Relationship) GetName() string {
	return r.name
}

// GetSourceEntity returns source entity
func (r *Relationship) GetSourceEntity() EntityInterface {
	return r.sourceEntity
}

// GetTargetEntity returns target entity
func (r *Relationship) GetTargetEntity() EntityInterface {
	return r.targetEntity
}

// GetSourceAttribute returns source attribute
func (r *Relationship) GetSourceAttribute() AttributeInterface {
	return r.sourceAttr
}

// GetTargetAttribute returns target attribute
func (r *Relationship) GetTargetAttribute() AttributeInterface {
	return r.targetAttr
}

// GetCardinality returns relationship cardinality (1:1, 1:N, N:1)
func (r *Relationship) GetCardinality() string {
	return r.cardinality
}

// IsOneToOne returns true if relationship is 1:1
func (r *Relationship) IsOneToOne() bool {
	return r.cardinality == OneToOne
}

// IsOneToMany returns true if relationship is 1:N
func (r *Relationship) IsOneToMany() bool {
	return r.cardinality == OneToMany
}

// IsManyToOne returns true if relationship is N:1
func (r *Relationship) IsManyToOne() bool {
	return r.cardinality == ManyToOne
}

// determineCardinality analyzes attributes to determine cardinality
func (r *Relationship) determineCardinality() {
	// Both are unique: one-to-one relationship
	if r.sourceAttr.IsUnique() && r.targetAttr.IsUnique() {
		r.cardinality = OneToOne
		return
	}

	// Source is unique, target is not: one-to-many relationship
	if r.sourceAttr.IsUnique() && !r.targetAttr.IsUnique() {
		r.cardinality = OneToMany
		return
	}

	// Target is unique, source is not: many-to-one relationship
	if !r.sourceAttr.IsUnique() && r.targetAttr.IsUnique() {
		r.cardinality = ManyToOne
		return
	}

	// Default to many-to-one, though this should not happen due to validation
	r.cardinality = ManyToOne
}

// GetTargetValueForSourceRow returns a target PK value for a specific source row
// Uses cardinality-appropriate distribution algorithms for realistic data
func (r *Relationship) GetTargetValueForSourceRow(sourceRowIndex int, autoCardinality bool) (string, error) {
	if r.targetEntity == nil {
		return "", fmt.Errorf("target entity is nil for relationship %s", r.id)
	}

	if r.sourceEntity == nil {
		return "", fmt.Errorf("source entity is nil for relationship %s", r.id)
	}

	targetRowCount := r.targetEntity.GetRowCount()
	if targetRowCount == 0 {
		return "", fmt.Errorf("target entity %s has no rows for relationship %s", r.targetEntity.GetName(), r.id)
	}

	sourceRowCount := r.sourceEntity.GetRowCount()
	if sourceRowCount == 0 {
		return "", fmt.Errorf("source entity %s has no rows for relationship %s", r.sourceEntity.GetName(), r.id)
	}

	// Select target index using appropriate algorithm
	targetIndex := r.selectTargetIndex(sourceRowIndex, targetRowCount, autoCardinality)

	// Get the target row and extract PK value
	targetRow := r.targetEntity.GetRowByIndex(targetIndex)
	if targetRow == nil {
		return "", fmt.Errorf("unable to get target row at index %d for relationship %s", targetIndex, r.id)
	}

	if r.targetAttr == nil {
		return "", fmt.Errorf("target attribute is nil for relationship %s", r.id)
	}

	targetValue := targetRow.GetValue(r.targetAttr.GetName())
	return targetValue, nil
}

// selectTargetIndex chooses the appropriate target index based on cardinality and settings
func (r *Relationship) selectTargetIndex(sourceRowIndex, targetRowCount int, autoCardinality bool) int {
	if !autoCardinality || r.cardinality == OneToOne {
		// Round-robin for predictable distribution or 1:1 unique assignment
		return sourceRowIndex % targetRowCount
	}

	// Use power law clustering for many-to-one relationships
	return r.powerLawIndex(sourceRowIndex, targetRowCount)
}

// powerLawIndex generates a power law distributed index for realistic clustering
func (r *Relationship) powerLawIndex(sourceRowIndex, targetCount int) int {
	if targetCount <= 1 {
		return 0
	}

	// Get the maximum source row index (dataVolume - 1) from the source entity
	maxSourceIndex := r.sourceEntity.GetRowCount() - 1
	if maxSourceIndex <= 0 {
		return 0
	}

	// DEBUG: Log when row count is suspiciously small during iteration
	if sourceRowIndex > 10 && maxSourceIndex < 10 {
		fmt.Fprintf(os.Stderr, "\nDEBUG: Suspicious! sourceRowIndex=%d but maxSourceIndex=%d (entity has %d rows)\n",
			sourceRowIndex, maxSourceIndex, r.sourceEntity.GetRowCount())
	}

	// Normalize sourceRowIndex to [0, 1] range
	normalizedIndex := float64(sourceRowIndex) / float64(maxSourceIndex)

	// Create relationship-specific alpha variation to avoid identical distributions
	baseAlpha := 1.3                               // Base power law exponent for moderate clustering
	alphaVariation := float64(len(r.id)%10) * 0.05 // Vary alpha by 0-0.45 based on relationship ID length
	alpha := baseAlpha + alphaVariation

	// Apply power law: y = x^alpha where x is normalized input
	// This creates clustering toward index 0 (lower indices more popular)
	powerValue := math.Pow(normalizedIndex, alpha)

	// Scale to target count range [0, targetCount-1]
	index := int(powerValue * float64(targetCount-1))

	return index
}
