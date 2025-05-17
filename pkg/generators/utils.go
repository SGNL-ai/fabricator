package generators

import (
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

// GetEntityFileName returns the filename for a given entity
// based on its external ID
func GetEntityFileName(csvData *models.CSVData) string {
	if csvData == nil {
		return "unknown"
	}

	// Handle both formats: with namespace prefix (e.g., "KeystoneV1/Entity") and without
	if strings.Contains(csvData.ExternalId, "/") {
		parts := strings.Split(csvData.ExternalId, "/")
		return parts[len(parts)-1] + ".csv"
	} else {
		// If no namespace prefix, just use the external ID
		return csvData.ExternalId + ".csv"
	}
}

// TODO: Refactor to use direct YAML model access rather than a pre-built map
// IsUniqueAttribute checks if an attribute is marked as uniqueId=true
func IsUniqueAttribute(entityID string, attrName string, uniqueIdMap map[string][]string) bool {
	uniqueAttrs, exists := uniqueIdMap[entityID]
	if !exists {
		return false
	}

	for _, uniqueAttr := range uniqueAttrs {
		if uniqueAttr == attrName {
			return true
		}
	}
	return false
}