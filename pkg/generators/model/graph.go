package model

import (
	"errors"
	"fmt"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/SGNL-ai/fabricator/pkg/util"
	"github.com/dominikbraun/graph"
)

// Error definitions for Graph operations
var (
	ErrNilYAMLModel         = errors.New("YAML model cannot be nil")
	ErrNoEntities           = errors.New("YAML model must contain at least one entity")
	ErrEntityNotFound       = errors.New("entity not found")
	ErrRelationshipNotFound = errors.New("relationship not found")
	ErrCircularDependency   = errors.New("circular dependency detected in entity relationships")
	ErrInvalidRelationship  = errors.New("invalid relationship definition")
)

// Graph represents the overall model including entities and relationships
type Graph struct {
	entities            map[string]EntityInterface         // Maps entity ID to Entity object
	entitiesList        []EntityInterface                  // Pre-computed list of all entities
	relationships       map[string]RelationshipInterface   // Maps relationship ID to Relationship object
	relationshipsList   []RelationshipInterface            // Pre-computed list of all relationships
	entityRelationships map[string][]RelationshipInterface // Maps entity ID to its relationships
	attributeToEntity   map[string]EntityInterface         // Maps attribute externalID to its containing entity
	yamlModel           *parser.SORDefinition              // Reference to original YAML model
}

// NewGraph creates a new Graph from the YAML model
// Following our four-step constructor pattern:
// 1. Object creation - Initialize the Graph with empty maps and lists
// 2. Validation - Validate the YAML model is not nil and has required elements
// 3. Setup - Create entities and relationships from YAML
// 4. Business logic - Validate graph integrity and build indexes
func NewGraph(yamlModel *parser.SORDefinition) (GraphInterface, error) {
	// 1. Object creation - Create a new Graph with initialized data structures
	graph := &Graph{
		entities:            make(map[string]EntityInterface),
		entitiesList:        make([]EntityInterface, 0),
		relationships:       make(map[string]RelationshipInterface),
		relationshipsList:   make([]RelationshipInterface, 0),
		entityRelationships: make(map[string][]RelationshipInterface),
		attributeToEntity:   make(map[string]EntityInterface),
		yamlModel:           yamlModel,
	}

	// 2. Validate the YAML model
	if yamlModel == nil {
		return nil, ErrNilYAMLModel
	}

	if len(yamlModel.Entities) == 0 {
		return nil, ErrNoEntities
	}

	// 3. Create entities and relationships from YAML
	if err := graph.createEntitiesFromYAML(yamlModel.Entities); err != nil {
		return nil, err
	}

	if err := graph.createRelationshipsFromYAML(yamlModel.Relationships); err != nil {
		return nil, err
	}

	// 4. Build optimized data structures for access
	graph.buildIndexes()

	return graph, nil
}

// GetEntity gets an entity by ID with existence check
func (g *Graph) GetEntity(id string) (EntityInterface, bool) {
	entity, exists := g.entities[id]
	return entity, exists
}

// GetAllEntities returns all entities in the graph
func (g *Graph) GetAllEntities() map[string]EntityInterface {
	return g.entities
}

// GetEntitiesList returns a slice of all entities in the graph
func (g *Graph) GetEntitiesList() []EntityInterface {
	// Convert to interface slice
	result := make([]EntityInterface, len(g.entitiesList))
	copy(result, g.entitiesList)
	return result
}

// GetRelationship gets a relationship by ID with existence check
func (g *Graph) GetRelationship(id string) (RelationshipInterface, bool) {
	relationship, exists := g.relationships[id]
	return relationship, exists
}

// GetAllRelationships returns all relationships in the graph
func (g *Graph) GetAllRelationships() []RelationshipInterface {
	return g.relationshipsList
}

// GetRelationshipsForEntity returns all relationships that involve a specific entity
func (g *Graph) GetRelationshipsForEntity(entityID string) []RelationshipInterface {
	return g.entityRelationships[entityID]
}

// GetTopologicalOrder returns entities in dependency order for generation
func (g *Graph) GetTopologicalOrder() ([]string, error) {

	// Use the existing YAML model directly
	entityGraph, err := util.BuildEntityDependencyGraph(g.yamlModel.Entities, g.yamlModel.Relationships, true)
	if err != nil {
		if errors.Is(err, graph.ErrEdgeCreatesCycle) {
			return nil, ErrCircularDependency
		}
		return nil, err
	}

	// Get topological order using the utility function
	order, err := util.GetTopologicalOrder(entityGraph)
	if err != nil {
		fmt.Printf("GetTopologicalOrder error: %v\n", err)
		return nil, err
	}
	return order, nil
}

// createEntitiesFromYAML creates Entity objects from YAML model definition
func (g *Graph) createEntitiesFromYAML(yamlEntities map[string]parser.Entity) error {
	// First, create all entities with their attributes
	for entityID, yamlEntity := range yamlEntities {
		// Convert YAML attributes to model attributes
		attributes := make([]AttributeInterface, 0, len(yamlEntity.Attributes))

		for _, yamlAttr := range yamlEntity.Attributes {
			// Create attribute with all necessary details
			attr := newAttribute(
				yamlAttr.Name,
				yamlAttr.ExternalId,
				yamlAttr.AttributeAlias,
				yamlAttr.Type,
				yamlAttr.UniqueId,
				yamlAttr.Description,
				nil, // Parent entity will be set by newEntity
			)
			attributes = append(attributes, attr)
		}

		// Create entity with attributes
		entity, err := newEntity(
			entityID,
			yamlEntity.ExternalId,
			yamlEntity.DisplayName,
			yamlEntity.Description,
			attributes,
			g,
		)

		if err != nil {
			return fmt.Errorf("failed to create entity %s: %w", entityID, err)
		}

		// Add entity to the graph
		g.entities[entityID] = entity
	}

	// Then build the attribute to entity lookup map
	for entityID, entity := range g.entities {
		yamlEntity := g.yamlModel.Entities[entityID]

		for _, yamlAttr := range yamlEntity.Attributes {
			g.attributeToEntity[fmt.Sprintf("%s.%s", entityID, yamlAttr.ExternalId)] = entity
			if yamlAttr.AttributeAlias != "" {
				g.attributeToEntity[yamlAttr.AttributeAlias] = entity
			}
		}
	}

	return nil
}

// createRelationshipsFromYAML creates Relationship objects from YAML model definition
func (g *Graph) createRelationshipsFromYAML(yamlRelationships map[string]parser.Relationship) error {
	// Iterate over the relationships in the YAML
	for relationshipID, yamlRel := range yamlRelationships {
		// Skip relationships defined with path (complex relationships)
		if len(yamlRel.Path) > 0 {
			// TODO: Handle path-based relationships when needed
			continue
		}

		// Get source entity from FromAttribute
		sourceEntity := g.attributeToEntity[yamlRel.FromAttribute]
		if sourceEntity == nil {
			return fmt.Errorf("source entity not found for relationship %s (attribute: %s)",
				relationshipID, yamlRel.FromAttribute)
		}

		// Get target entity from ToAttribute
		targetEntity := g.attributeToEntity[yamlRel.ToAttribute]
		if targetEntity == nil {
			return fmt.Errorf("target entity not found for relationship %s (attribute: %s)",
				relationshipID, yamlRel.ToAttribute)
		}

		// Call addRelationship on the source entity with the external IDs
		// let addRelationship handle finding the actual attribute names
		relationship, err := sourceEntity.addRelationship(
			relationshipID,
			yamlRel.Name,
			targetEntity,
			yamlRel.FromAttribute, //this could be dotted notation or attributeAlias
			yamlRel.ToAttribute,
		)

		if err != nil {
			return fmt.Errorf("failed to create relationship %s: %w", relationshipID, err)
		}

		// Only add non-nil relationships to the graph
		if relationship != nil {
			g.relationships[relationshipID] = relationship
		}
	}

	return nil
}

// buildIndexes builds all the optimized data structures for faster lookups
func (g *Graph) buildIndexes() {
	// Clear existing indexes
	g.entitiesList = make([]EntityInterface, 0, len(g.entities))
	g.relationshipsList = make([]RelationshipInterface, 0, len(g.relationships))
	g.entityRelationships = make(map[string][]RelationshipInterface)

	// Build entities list
	for _, entity := range g.entities {
		g.entitiesList = append(g.entitiesList, entity)

		// Initialize empty relationship list for each entity
		g.entityRelationships[entity.GetID()] = make([]RelationshipInterface, 0)
	}

	// Build relationships list and entity relationships map
	for _, rel := range g.relationships {
		g.relationshipsList = append(g.relationshipsList, rel)

		// Add to source entity's relationships
		sourceEntityID := rel.GetSourceEntity().GetID()
		g.entityRelationships[sourceEntityID] = append(g.entityRelationships[sourceEntityID], rel)

		// Add to target entity's relationships if different from source
		targetEntityID := rel.GetTargetEntity().GetID()
		if targetEntityID != sourceEntityID {
			g.entityRelationships[targetEntityID] = append(g.entityRelationships[targetEntityID], rel)
		}
	}
}

