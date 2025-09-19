package pipeline

import (
	"fmt"
	"strconv"
	"time"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/brianvoe/gofakeit/v6"
)

// FieldGenerator handles generation of non-ID and non-relationship fields
type FieldGenerator struct {
	// Configuration options can be added here
}

// NewFieldGenerator creates a new field generator
func NewFieldGenerator() FieldGeneratorInterface {
	return &FieldGenerator{}
}

// GenerateFields generates values for all non-ID and non-relationship fields
func (g *FieldGenerator) GenerateFields(graph *model.Graph) error {
	if graph == nil {
		return fmt.Errorf("graph cannot be nil")
	}

	// Process each entity
	for _, entity := range graph.GetAllEntities() {
		// Show progress for current entity (will be cleared)
		fmt.Printf("\r%-80s\r→ Generating fields for %s...", "", entity.GetName())

		// Get non-ID, non-relationship attributes that need values
		fieldsToGenerate := entity.GetNonRelationshipAttributes()

		// Filter out unique attributes (already handled by ID generator)
		var regularFields []model.AttributeInterface
		for _, attr := range fieldsToGenerate {
			if !attr.IsUnique() {
				regularFields = append(regularFields, attr)
			}
		}

		// Skip if no regular fields to generate
		if len(regularFields) == 0 {
			continue
		}

		// Use iterator to set field values in entity rows
		err := entity.ForEachRow(func(row *model.Row) error {
			for _, attr := range regularFields {
				// Generate appropriate value based on attribute type and name
				value := g.generateFieldValue(attr)
				row.SetValue(attr.GetName(), value)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to generate fields for entity %s: %w", entity.GetExternalID(), err)
		}
	}

	// Clear field generation progress line
	fmt.Printf("\r%-80s\r", "")

	return nil
}

// generateFieldValue generates an appropriate value for an attribute
func (g *FieldGenerator) generateFieldValue(attr model.AttributeInterface) string {
	attrName := attr.GetName()
	dataType := attr.GetDataType()

	// Generate based on field name patterns first
	switch {
	case contains(attrName, "email"):
		return gofakeit.Email()
	case contains(attrName, "name"):
		return gofakeit.Name()
	case contains(attrName, "phone"):
		return gofakeit.Phone()
	case contains(attrName, "address"):
		return gofakeit.Address().Address
	case contains(attrName, "status"):
		return gofakeit.RandomString([]string{"active", "inactive", "pending"})
	case contains(attrName, "date"), contains(attrName, "time"):
		return gofakeit.Date().Format(time.RFC3339)
	}

	// Generate based on data type
	switch dataType {
	case "Integer", "Int64":
		return strconv.Itoa(gofakeit.Number(1, 1000))
	case "Boolean", "Bool":
		return strconv.FormatBool(gofakeit.Bool())
	case "Date":
		return gofakeit.Date().Format("2006-01-02")
	case "DateTime":
		return gofakeit.Date().Format(time.RFC3339)
	case "Float", "Double":
		return fmt.Sprintf("%.2f", gofakeit.Float64Range(1.0, 100.0))
	default:
		// Default to string
		return gofakeit.Word()
	}
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
