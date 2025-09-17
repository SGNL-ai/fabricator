package generators

import (
	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/SGNL-ai/fabricator/pkg/util"
)

// RelationshipType represents the type of relationship between entities
type RelationshipType int

const (
	// RelationshipOneToOne represents a 1:1 relationship (default case)
	RelationshipOneToOne RelationshipType = iota
	// RelationshipOneToMany represents a 1:N relationship
	RelationshipOneToMany
	// RelationshipManyToOne represents an N:1 relationship
	RelationshipManyToOne
)

// RelationshipContext contains all the information needed to process a relationship
type RelationshipContext struct {
	// Entity information
	FromEntityID string
	ToEntityID   string
	FromData     *models.CSVData
	ToData       *models.CSVData

	// Attribute information
	FromAttrName  string
	ToAttrName    string
	FromAttrIndex int
	ToAttrIndex   int
	FromIsUnique  bool
	ToIsUnique    bool

	// Relationship metadata
	RelationshipType RelationshipType
	AutoCardinality  bool
}

// NewRelationshipContext creates a new context for processing a relationship
func NewRelationshipContext(g *CSVGenerator, fromEntityID string, link models.RelationshipLink) *RelationshipContext {
	ctx := &RelationshipContext{
		FromEntityID:    fromEntityID,
		ToEntityID:      link.ToEntityID,
		FromAttrName:    link.FromAttribute,
		ToAttrName:      link.ToAttribute,
		AutoCardinality: g.AutoCardinality,
	}

	// Get entity data
	ctx.FromData = g.EntityData[fromEntityID]
	ctx.ToData = g.EntityData[link.ToEntityID]

	if ctx.FromData == nil || ctx.ToData == nil {
		return nil
	}

	// Find attribute indices
	ctx.FromAttrIndex = findAttributeIndex(ctx.FromData.Headers, link.FromAttribute)
	ctx.ToAttrIndex = findAttributeIndex(ctx.ToData.Headers, link.ToAttribute)

	if ctx.FromAttrIndex == -1 || ctx.ToAttrIndex == -1 {
		return nil
	}

	// Determine attribute uniqueness
	ctx.FromIsUnique = IsUniqueAttribute(fromEntityID, ctx.FromAttrName, g.uniqueIdAttributes)
	ctx.ToIsUnique = IsUniqueAttribute(link.ToEntityID, ctx.ToAttrName, g.uniqueIdAttributes)

	// Determine relationship type based solely on uniqueness constraints
	ctx.RelationshipType = determineRelationshipType(ctx.FromIsUnique, ctx.ToIsUnique)

	return ctx
}

// findAttributeIndex locates the index of an attribute in the headers using exact name matching
func findAttributeIndex(headers []string, attrName string) int {
	for i, header := range headers {
		// Only use direct exact matching - no heuristics
		if header == attrName {
			return i
		}
	}

	return -1
}

// determineRelationshipType analyzes the uniqueness constraints to determine relationship type
func determineRelationshipType(fromIsUnique, toIsUnique bool) RelationshipType {
	// Use uniqueness as the sole determinant of relationship type
	if fromIsUnique && !toIsUnique {
		// From is unique, To is not: 1:N relationship
		return RelationshipOneToMany
	} else if !fromIsUnique && toIsUnique {
		// From is not unique, To is: N:1 relationship
		return RelationshipManyToOne
	}

	// Default: both unique or both non-unique -> 1:1 relationship
	return RelationshipOneToOne
}

// handleRelationship processes a relationship based on its type
func (g *CSVGenerator) handleRelationship(ctx *RelationshipContext) {
	switch ctx.RelationshipType {
	case RelationshipOneToMany:
		g.handleOneToManyRelationship(ctx)
	case RelationshipManyToOne:
		g.handleManyToOneRelationship(ctx)
	default:
		g.handleOneToOneRelationship(ctx)
	}
}

// handleOneToManyRelationship handles one-to-many relationships
// In 1:N relationships, FromEntity has unique value (e.g., id) that ToEntity refers to many times
func (g *CSVGenerator) handleOneToManyRelationship(ctx *RelationshipContext) {
	// Collect all values from the from entity (primary key values)
	fromValues := collectNonEmptyValues(ctx.FromData.Rows, ctx.FromAttrIndex)

	// If no values in from entity, we can't make them consistent
	if len(fromValues) == 0 {
		return
	}

	// For 1:N relationships, the many side (ToEntity) should reference values from the one side (FromEntity)
	// Update the foreign key in the ToEntity to reference primary keys from FromEntity
	for i, row := range ctx.ToData.Rows {
		if ctx.ToAttrIndex < len(row) {
			// Distribute values from primary entity evenly among secondary entities
			fromValueIdx := i % len(fromValues)
			row[ctx.ToAttrIndex] = fromValues[fromValueIdx]
			ctx.ToData.Rows[i] = row
		}
	}
}

// handleManyToOneRelationship handles many-to-one relationships
// This often involves clustering multiple "from" entity rows to the same "to" entity value
func (g *CSVGenerator) handleManyToOneRelationship(ctx *RelationshipContext) {
	// Collect all values from the "to" entity
	toValues := collectNonEmptyValues(ctx.ToData.Rows, ctx.ToAttrIndex)

	// If no values in "to" entity, we can't make them consistent
	if len(toValues) == 0 {
		return
	}

	// Create clusters of values
	numClusters := len(toValues)
	if numClusters > 0 {
		// With auto-cardinality, use fewer clusters to create more visible grouping
		if ctx.AutoCardinality && numClusters > 2 {
			// For stronger many-to-one effect, use just 2-3 clusters
			numClusters = 2 + util.CryptoRandInt(2) // 2-3 clusters
		}

		clusterSize := len(ctx.FromData.Rows) / numClusters
		if clusterSize < 1 {
			clusterSize = 1
		}

		for i, row := range ctx.FromData.Rows {
			// Determine which cluster this row belongs to
			clusterIndex := i / clusterSize
			if clusterIndex >= numClusters {
				clusterIndex = numClusters - 1
			}

			// Assign the value from the cluster
			row[ctx.FromAttrIndex] = toValues[clusterIndex]
			ctx.FromData.Rows[i] = row
		}
	}
}

// handleOneToOneRelationship handles one-to-one relationships
func (g *CSVGenerator) handleOneToOneRelationship(ctx *RelationshipContext) {
	// Collect all values from the "to" entity
	toValues := collectNonEmptyValues(ctx.ToData.Rows, ctx.ToAttrIndex)

	// If no values in "to" entity, we can't make them consistent
	if len(toValues) == 0 {
		return
	}

	if ctx.FromIsUnique {
		// For unique attributes, ensure we don't reuse values
		// If we have more rows than unique values, we need to generate more unique values
		for i, row := range ctx.FromData.Rows {
			if i < len(toValues) {
				// Try to use each target value once
				row[ctx.FromAttrIndex] = toValues[i]
			} else {
				// We've used all available target values, so generate new ones
				row[ctx.FromAttrIndex] = g.ensureUniqueValue(ctx.FromEntityID, ctx.FromAttrName, "")
			}
			ctx.FromData.Rows[i] = row
		}
	} else {
		// For non-unique attributes, just assign randomly
		for i, row := range ctx.FromData.Rows {
			// Use a random value from the "to" entity for each row in the "from" entity
			row[ctx.FromAttrIndex] = toValues[util.CryptoRandInt(len(toValues))]
			ctx.FromData.Rows[i] = row
		}
	}
}

// Helper function to collect non-empty values from a specific column
func collectNonEmptyValues(rows [][]string, columnIndex int) []string {
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		// Only add non-empty values
		if columnIndex < len(row) && row[columnIndex] != "" {
			values = append(values, row[columnIndex])
		}
	}
	return values
}
