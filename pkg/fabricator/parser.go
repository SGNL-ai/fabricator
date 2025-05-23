package fabricator

import (
	"fmt"
	"os"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"gopkg.in/yaml.v3"
)

// Parser handles the parsing of YAML definition files
type Parser struct {
	Definition *models.SORDefinition
	FilePath   string
}

// NewParser creates a new Parser instance
func NewParser(filePath string) *Parser {
	return &Parser{
		FilePath: filePath,
	}
}

// Parse loads and parses the YAML file
func (p *Parser) Parse() error {
	// Read the YAML file
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the YAML content
	p.Definition = &models.SORDefinition{}
	err = yaml.Unmarshal(data, p.Definition)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed data
	err = p.validate()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// validate checks if the parsed YAML has valid structure
func (p *Parser) validate() error {
	if p.Definition == nil {
		return fmt.Errorf("empty definition")
	}

	if len(p.Definition.Entities) == 0 {
		return fmt.Errorf("no entities defined")
	}

	// Check that each entity has at least one attribute and a valid external ID
	for id, entity := range p.Definition.Entities {
		if entity.ExternalId == "" {
			return fmt.Errorf("entity %s missing externalId", id)
		}

		if len(entity.Attributes) == 0 {
			return fmt.Errorf("entity %s has no attributes", id)
		}

		// Check for uniqueId attribute and ensure only one attribute is marked as uniqueId
		hasUniqueId := false
		uniqueIdAttrs := []string{}
		for _, attr := range entity.Attributes {
			if attr.UniqueId {
				hasUniqueId = true
				uniqueIdAttrs = append(uniqueIdAttrs, attr.Name)
			}
		}

		if !hasUniqueId {
			return fmt.Errorf("entity %s (%s) has no attribute marked as uniqueId",
				id, entity.DisplayName)
		}

		// Check that at least one attribute is marked as uniqueId (already checked above)
		// But this code is added for clarity and future-proofing
		if len(uniqueIdAttrs) == 0 {
			return fmt.Errorf("entity %s (%s) has no attribute marked as uniqueId",
				id, entity.DisplayName)
		}
	}

	// Basic validation for relationships
	for id, rel := range p.Definition.Relationships {
		// Skip path-based relationships for validation
		if len(rel.Path) > 0 {
			continue
		}

		if rel.FromAttribute == "" {
			return fmt.Errorf("relationship %s missing fromAttribute", id)
		}

		if rel.ToAttribute == "" {
			return fmt.Errorf("relationship %s missing toAttribute", id)
		}
	}

	return nil
}

// GetEntityByExternalId returns an entity by its external ID
func (p *Parser) GetEntityByExternalId(externalId string) (*models.Entity, string, error) {
	for id, entity := range p.Definition.Entities {
		if entity.ExternalId == externalId {
			return &entity, id, nil
		}
	}
	return nil, "", fmt.Errorf("entity with externalId %s not found", externalId)
}

// GetEntityById returns an entity by its ID
func (p *Parser) GetEntityById(id string) (*models.Entity, error) {
	entity, exists := p.Definition.Entities[id]
	if !exists {
		return nil, fmt.Errorf("entity with id %s not found", id)
	}
	return &entity, nil
}

// GetCSVFilenames returns a map of external IDs to CSV filenames
func (p *Parser) GetCSVFilenames() map[string]string {
	filenames := make(map[string]string)

	for _, entity := range p.Definition.Entities {
		// Get the base name for the CSV file from the external ID
		parts := strings.Split(entity.ExternalId, "/")
		baseFilename := parts[len(parts)-1]

		filenames[entity.ExternalId] = baseFilename + ".csv"
	}

	return filenames
}

// FindRelationshipsForEntity finds all relationships involving a given entity
func (p *Parser) FindRelationshipsForEntity(entityId string) map[string]models.Relationship {
	// Get the entity
	entity, err := p.GetEntityById(entityId)
	if err != nil {
		return nil
	}

	// Create a map of attribute aliases for this entity
	attributeAliases := make(map[string]bool)
	for _, attr := range entity.Attributes {
		attributeAliases[attr.AttributeAlias] = true
	}

	// Find relationships that involve this entity
	result := make(map[string]models.Relationship)
	for id, rel := range p.Definition.Relationships {
		// Skip path-based relationships for now
		if len(rel.Path) > 0 {
			continue
		}

		// Check if this entity is involved in the relationship
		if attributeAliases[rel.FromAttribute] || attributeAliases[rel.ToAttribute] {
			result[id] = rel
		}
	}

	return result
}

// FindEntityRelationships finds direct relationship links for a given entity
func (p *Parser) FindEntityRelationships(entityId string) []models.RelationshipLink {
	relationships := []models.RelationshipLink{}
	entity, err := p.GetEntityById(entityId)
	if err != nil {
		return relationships
	}

	// Create a map of all attribute aliases to their entities and attribute names
	attributeAliasMap := make(map[string]struct {
		EntityId      string
		AttributeName string
	})

	for id, e := range p.Definition.Entities {
		for _, attr := range e.Attributes {
			attributeAliasMap[attr.AttributeAlias] = struct {
				EntityId      string
				AttributeName string
			}{
				EntityId:      id,
				AttributeName: attr.Name,
			}
		}
	}

	// Get the attribute aliases for this entity
	entityAttrAliases := make(map[string]string) // alias -> attr name
	for _, attr := range entity.Attributes {
		entityAttrAliases[attr.AttributeAlias] = attr.Name
	}

	// Find direct relationships
	for _, rel := range p.Definition.Relationships {
		// Skip path-based relationships
		if len(rel.Path) > 0 {
			continue
		}

		// Check if this entity's attribute is the "from" in a relationship
		if fromAttrName, ok := entityAttrAliases[rel.FromAttribute]; ok {
			// Find the "to" entity and attribute
			if toInfo, ok := attributeAliasMap[rel.ToAttribute]; ok {
				link := models.RelationshipLink{
					FromEntityID:  entityId,
					ToEntityID:    toInfo.EntityId,
					FromAttribute: fromAttrName,
					ToAttribute:   toInfo.AttributeName,
				}
				relationships = append(relationships, link)
			}
		}

		// Check if this entity's attribute is the "to" in a relationship
		if toAttrName, ok := entityAttrAliases[rel.ToAttribute]; ok {
			// Find the "from" entity and attribute
			if fromInfo, ok := attributeAliasMap[rel.FromAttribute]; ok {
				link := models.RelationshipLink{
					FromEntityID:  fromInfo.EntityId,
					ToEntityID:    entityId,
					FromAttribute: fromInfo.AttributeName,
					ToAttribute:   toAttrName,
				}
				relationships = append(relationships, link)
			}
		}
	}

	return relationships
}

// GetUniqueIdAttributeFor returns the attribute that's marked as the unique ID for an entity
func (p *Parser) GetUniqueIdAttributeFor(entityId string) (*models.Attribute, error) {
	entity, err := p.GetEntityById(entityId)
	if err != nil {
		return nil, err
	}

	for _, attr := range entity.Attributes {
		if attr.UniqueId {
			return &attr, nil
		}
	}

	return nil, fmt.Errorf("no uniqueId attribute found for entity %s", entityId)
}

// GetNamespacePrefix gets the common prefix used in external IDs
func (p *Parser) GetNamespacePrefix() string {
	for _, entity := range p.Definition.Entities {
		parts := strings.Split(entity.ExternalId, "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}
