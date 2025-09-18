package fabricator

import (
	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/fatih/color"
)

// PrintGraphStatistics displays detailed statistics about the parsed graph
func PrintGraphStatistics(stats *model.GraphStatistics) {
	color.Green("✓ Successfully parsed YAML definition")
	color.Green("  SOR name: %s", stats.SORName)
	color.Green("  Description: %s", stats.Description)

	// Display namespace format
	if len(stats.NamespaceFormats) == 1 {
		for prefix, count := range stats.NamespaceFormats {
			if prefix == "(no namespace)" {
				color.Green("  Namespace format detected: No namespace prefix (%d entities)", count)
			} else {
				color.Green("  Namespace format detected: %s/EntityName (%d entities)", prefix, count)
			}
		}
	} else {
		color.Green("  Namespace formats detected: multiple")
		for prefix, count := range stats.NamespaceFormats {
			color.Green("    %s: %d entities", prefix, count)
		}
	}

	color.Green("  Entities: %d", stats.EntityCount)
	color.Green("  Total attributes: %d", stats.TotalAttributes)
	color.Green("     - Unique ID attributes: %d", stats.UniqueAttributes)
	color.Green("     - Indexed attributes: %d", stats.IndexedAttributes)
	color.Green("     - List attributes: %d", stats.ListAttributes)
	color.Green("  Relationships: %d total (%d direct, %d path-based)",
		stats.RelationshipCount, stats.DirectRelationships, stats.PathBasedRelationships)
}
