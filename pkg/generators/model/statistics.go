package model

import (
	"strings"
)

// GraphStatistics contains statistical information about a parsed graph
type GraphStatistics struct {
	SORName                string
	Description            string
	EntityCount            int
	TotalAttributes        int
	UniqueAttributes       int
	IndexedAttributes      int
	ListAttributes         int
	RelationshipCount      int
	DirectRelationships    int
	PathBasedRelationships int
	NamespaceFormats       map[string]int
}

// GetStatistics returns comprehensive statistics about the graph
func (g *Graph) GetStatistics() *GraphStatistics {
	stats := &GraphStatistics{
		SORName:          g.yamlModel.DisplayName,
		Description:      g.yamlModel.Description,
		NamespaceFormats: make(map[string]int),
	}

	// Count entities
	stats.EntityCount = len(g.entities)

	// Analyze entities and attributes
	for _, entity := range g.entities {
		// Analyze namespace format from external ID
		externalID := entity.GetExternalID()
		if strings.Contains(externalID, "/") {
			parts := strings.Split(externalID, "/")
			if len(parts) >= 2 {
				prefix := parts[0]
				stats.NamespaceFormats[prefix]++
			}
		} else {
			stats.NamespaceFormats["(no namespace)"]++
		}

		// Count attributes
		for _, attr := range entity.GetAttributes() {
			stats.TotalAttributes++
			if attr.IsUnique() {
				stats.UniqueAttributes++
			}
			// Note: Indexed and List info would need to be stored in the attribute model
			// For now, we'll get these from the original YAML if needed
		}
	}

	// Get indexed and list attributes from original YAML (since model doesn't store these flags)
	for _, yamlEntity := range g.yamlModel.Entities {
		for _, yamlAttr := range yamlEntity.Attributes {
			if yamlAttr.Indexed {
				stats.IndexedAttributes++
			}
			if yamlAttr.List {
				stats.ListAttributes++
			}
		}
	}

	// Count relationships
	stats.RelationshipCount = len(g.relationships)

	// Analyze relationship types from original YAML
	for _, yamlRel := range g.yamlModel.Relationships {
		if len(yamlRel.Path) > 0 {
			stats.PathBasedRelationships++
		} else {
			stats.DirectRelationships++
		}
	}

	return stats
}
