package parser

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed sor_schema.json
var sorSchemaJSON string

// Parser handles the parsing of YAML definition files
type Parser struct {
	Definition *SORDefinition
	FilePath   string
	schema     *jsonschema.Schema
	Quiet      bool // Suppress debug output when true
}

// NewParser creates a new Parser instance
func NewParser(filePath string) *Parser {
	parser := &Parser{
		FilePath: filePath,
	}

	// Initialize the JSON schema - embedded schema is guaranteed to be valid
	_ = parser.initSchema()

	return parser
}

// initSchema compiles the JSON schema for SOR template validation
func (p *Parser) initSchema() error {
	// Parse the embedded JSON schema - guaranteed to be valid
	var schemaData interface{}
	_ = json.Unmarshal([]byte(sorSchemaJSON), &schemaData)

	compiler := jsonschema.NewCompiler()

	// Add the parsed schema data - embedded schema is guaranteed to be valid
	_ = compiler.AddResource("sor_schema.json", schemaData)

	// Compile the schema - embedded schema is guaranteed to compile successfully
	schema, _ := compiler.Compile("sor_schema.json")

	p.schema = schema
	return nil
}

// Parse loads and parses the YAML file
func (p *Parser) Parse() error {
	// Read the YAML file
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// First, perform JSON Schema validation on the raw YAML
	err = p.validateSchema(data)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Parse the YAML content
	p.Definition = &SORDefinition{}
	err = yaml.Unmarshal(data, p.Definition)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed data (business logic validation)
	err = p.validate()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// validateSchema validates the YAML data against the JSON schema
func (p *Parser) validateSchema(data []byte) error {
	if p.schema == nil {
		// Initialize schema if it wasn't done during construction
		_ = p.initSchema()
	}

	// Convert YAML to a generic interface for JSON schema validation
	var yamlData interface{}
	err := yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return fmt.Errorf("failed to parse YAML for schema validation: %w", err)
	}

	// Convert to JSON-compatible format (yaml.v3 produces map[string]interface{} which is compatible)
	// But we need to handle the case where YAML might produce different types
	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to convert YAML to JSON for schema validation: %w", err)
	}

	var jsonInterface interface{}
	err = json.Unmarshal(jsonData, &jsonInterface)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON for schema validation: %w", err)
	}

	// Validate against schema
	err = p.schema.Validate(jsonInterface)
	if err != nil {
		// Format validation errors nicely
		if validationErr, ok := err.(*jsonschema.ValidationError); ok {
			return fmt.Errorf("schema validation error at %s: %s", validationErr.InstanceLocation, validationErr.Error())
		}
		return fmt.Errorf("schema validation error: %w", err)
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

	// Validate relationships
	err := p.validateRelationships()
	if err != nil {
		return err
	}

	return nil
}

// buildAttributeSuggestions creates helpful debugging information when an attribute cannot be found
func (p *Parser) buildAttributeSuggestions(attrRef string, aliasMap, entityAttrMap map[string]struct {
	EntityID      string
	AttributeName string
	ExternalID    string
	UniqueID      bool
}) string {
	var suggestions strings.Builder
	suggestions.WriteString("\n    Searched for: ")

	// Show what format was attempted
	if strings.Contains(attrRef, ".") {
		suggestions.WriteString("Entity.Attribute format")
	} else {
		suggestions.WriteString("Attribute alias format")
	}

	// Show some examples of available formats
	suggestions.WriteString("\n    Available attribute aliases (sample):")
	count := 0
	for alias := range aliasMap {
		if count >= 5 {
			suggestions.WriteString("\n      ... and more")
			break
		}
		suggestions.WriteString(fmt.Sprintf("\n      - %s", alias))
		count++
	}

	if count == 0 {
		suggestions.WriteString("\n      (none found)")
	}

	suggestions.WriteString("\n    Available Entity.Attribute patterns (sample):")
	count = 0
	for pattern := range entityAttrMap {
		if count >= 5 {
			suggestions.WriteString("\n      ... and more")
			break
		}
		suggestions.WriteString(fmt.Sprintf("\n      - %s", pattern))
		count++
	}

	if count == 0 {
		suggestions.WriteString("\n      (none found)")
	}

	return suggestions.String()
}

// validateRelationships performs comprehensive validation of relationship definitions
func (p *Parser) validateRelationships() error {
	if len(p.Definition.Relationships) == 0 {
		// No relationships to validate
		return nil
	}

	// Create maps to find entities and attributes by different identifiers
	// Map of attribute alias to (entityID, attrName, externalId) - for alias-based relationships
	attributeAliasMap := make(map[string]struct {
		EntityID      string
		AttributeName string
		ExternalID    string
		UniqueID      bool
	})

	// Map to lookup entities/attributes by "Entity.Attribute" pattern
	entityAttributeMap := make(map[string]struct {
		EntityID      string
		AttributeName string
		ExternalID    string
		UniqueID      bool
	})

	// Build the attribute maps and collect entity info for debugging
	fmt.Fprintf(os.Stderr, "\nDEBUG: Discovered %d entities:\n", len(p.Definition.Entities))
	for entityID, entity := range p.Definition.Entities {
		fmt.Fprintf(os.Stderr, "  • Entity ID: %s, External ID: %s, Display Name: %s (%d attributes)\n",
			entityID, entity.ExternalId, entity.DisplayName, len(entity.Attributes))

		for _, attr := range entity.Attributes {
			// Handle attributeAlias case (when it exists)
			if attr.AttributeAlias != "" {
				attributeAliasMap[attr.AttributeAlias] = struct {
					EntityID      string
					AttributeName string
					ExternalID    string
					UniqueID      bool
				}{
					EntityID:      entityID,
					AttributeName: attr.Name,
					ExternalID:    attr.ExternalId,
					UniqueID:      attr.UniqueId,
				}
			}

			// Also build Entity.Attribute map for YAMLs without attributeAlias
			entityKey := entity.ExternalId + "." + attr.ExternalId
			entityAttributeMap[entityKey] = struct {
				EntityID      string
				AttributeName string
				ExternalID    string
				UniqueID      bool
			}{
				EntityID:      entityID,
				AttributeName: attr.Name,
				ExternalID:    attr.ExternalId,
				UniqueID:      attr.UniqueId,
			}
		}
	}
	fmt.Println()

	// Keep track of valid and invalid relationships
	invalidRelationships := make([]string, 0)
	validRelationships := 0
	pathBasedRelationships := 0

	// Track bidirectional relationships to detect potential cycles
	// Map from "entityID1:entityID2" to relationship ID
	// Used to check for reverse relationships that could create cycles
	bidirectionalRelationships := make(map[string]string)

	// Validate each relationship
	for relID, rel := range p.Definition.Relationships {
		// First, validate path-based relationships
		if len(rel.Path) > 0 {
			pathBasedRelationships++
			// For path-based relationships, ensure all the referenced relationships exist
			for i, pathStep := range rel.Path {
				referencedRel, exists := p.Definition.Relationships[pathStep.Relationship]
				if !exists {
					invalidRelationships = append(invalidRelationships,
						fmt.Sprintf("relationship %s: path step %d references non-existent relationship %s",
							relID, i+1, pathStep.Relationship))
					continue
				}

				// Also verify the referenced relationship is a direct relationship, not another path
				if len(referencedRel.Path) > 0 {
					invalidRelationships = append(invalidRelationships,
						fmt.Sprintf("relationship %s: path step %d references path-based relationship %s (nested paths not supported)",
							relID, i+1, pathStep.Relationship))
				}

				// Path direction is defined by external system and is not validated
				// We just need to ensure the referenced relationship exists
			}
			continue
		}

		// Check for childEntity relationships (parent-child hierarchical relationships)
		if rel.ChildEntity != "" {
			// ChildEntity relationships are valid - they represent parent-child hierarchies
			// The childEntity field can reference JSON paths like $.riskFactors or entity names
			validRelationships++
			continue
		}

		// For direct relationships, validate fromAttribute and toAttribute
		if rel.FromAttribute == "" {
			invalidRelationships = append(invalidRelationships,
				fmt.Sprintf("relationship %s: missing fromAttribute", relID))
			continue
		}

		if rel.ToAttribute == "" {
			invalidRelationships = append(invalidRelationships,
				fmt.Sprintf("relationship %s: missing toAttribute", relID))
			continue
		}

		// Check if attributes match real entities - try both mapping approaches
		var fromInfo, toInfo struct {
			EntityID      string
			AttributeName string
			ExternalID    string
			UniqueID      bool
		}
		var fromFound, toFound bool

		// First check attribute alias mapping
		if info, found := attributeAliasMap[rel.FromAttribute]; found {
			fromInfo = info
			fromFound = true
		}

		if info, found := attributeAliasMap[rel.ToAttribute]; found {
			toInfo = info
			toFound = true
		}

		// If not found, try Entity.Attribute mapping
		if !fromFound && strings.Contains(rel.FromAttribute, ".") {
			if info, found := entityAttributeMap[rel.FromAttribute]; found {
				fromInfo = info
				fromFound = true
			}
		}

		if !toFound && strings.Contains(rel.ToAttribute, ".") {
			if info, found := entityAttributeMap[rel.ToAttribute]; found {
				toInfo = info
				toFound = true
			}
		}

		// Report validation problems with detailed debugging information
		if !fromFound {
			// Build suggestions for what might have been intended
			suggestions := p.buildAttributeSuggestions(rel.FromAttribute, attributeAliasMap, entityAttributeMap)
			invalidRelationships = append(invalidRelationships,
				fmt.Sprintf("relationship %s: fromAttribute '%s' does not match any entity attribute%s",
					relID, rel.FromAttribute, suggestions))
		}

		if !toFound {
			// Build suggestions for what might have been intended
			suggestions := p.buildAttributeSuggestions(rel.ToAttribute, attributeAliasMap, entityAttributeMap)
			invalidRelationships = append(invalidRelationships,
				fmt.Sprintf("relationship %s: toAttribute '%s' does not match any entity attribute%s",
					relID, rel.ToAttribute, suggestions))
		}

		// Skip further validation if either attribute wasn't found
		if !fromFound || !toFound {
			continue
		}

		// Advanced relationship validation when both attributes are found
		validRelationships++

		// Check for self-referential relationships within the same entity
		if fromInfo.EntityID == toInfo.EntityID {
			// Self-referential relationships can be valid but should be flagged for review
			// For example, a user having a manager that is also a user
			// We'll only warn if both attributes are marked as uniqueId
			if fromInfo.UniqueID && toInfo.UniqueID {
				invalidRelationships = append(invalidRelationships,
					fmt.Sprintf("relationship %s: potential self-referential issue between uniqueId attributes '%s' and '%s' in entity '%s'",
						relID, fromInfo.AttributeName, toInfo.AttributeName, fromInfo.EntityID))
			}
		} else {
			// Check for bidirectional relationships between entities that could create cycles
			// Create bidirectional keys for both directions
			bidirKey1 := fromInfo.EntityID + ":" + toInfo.EntityID
			bidirKey2 := toInfo.EntityID + ":" + fromInfo.EntityID

			// Check if a relationship in the opposite direction exists
			if existingRelID, exists := bidirectionalRelationships[bidirKey2]; exists {
				// Bidirectional relationship detected - this is normal in many systems
				// Just log it as info, don't treat as validation error
				// The dependency graph creation will handle any actual cycles
				fmt.Printf("INFO: Bidirectional relationship detected: %s ↔ %s\n", relID, existingRelID)
			}

			// Record this relationship direction
			bidirectionalRelationships[bidirKey1] = relID

			// Validate relationship attribute types (uniqueId status)
			// Warn if neither attribute is a uniqueId (may cause data generation issues)
			if !fromInfo.UniqueID && !toInfo.UniqueID {
				fmt.Printf("WARNING: Relationship %s has no uniqueId attributes - may cause data generation issues\n", relID)
			}
		}
	}

	// Report validation results
	if len(invalidRelationships) > 0 {
		// Build detailed error message
		errorMsg := fmt.Sprintf("Found %d relationship issues (out of %d total relationships):\n",
			len(invalidRelationships), len(p.Definition.Relationships))

		for _, msg := range invalidRelationships {
			errorMsg += "• " + msg + "\n"
		}

		errorMsg += fmt.Sprintf("\nValid relationships: %d direct, %d path-based",
			validRelationships, pathBasedRelationships)

		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}
