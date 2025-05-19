# Making Your Graph Model Mockable with Testify

## Introduction

This guide outlines the changes needed to make your Graph model more testable and mockable using the testify package. These changes focus on:

1. Creating interfaces for key components
2. Updating constructors to return interfaces
3. Improving dependency injection
4. Setting up testify mocks

## Step 1: Define Interfaces for Key Components

First, define interfaces for your key types (Graph, Entity, Relationship, etc.):

```go
// GraphInterface defines the operations that can be performed on a Graph
type GraphInterface interface {
    GetEntity(id string) (*Entity, bool)
    GetAllEntities() map[string]*Entity
    GetEntitiesList() []*Entity
    GetRelationship(id string) (*Relationship, bool)
    GetAllRelationships() []*Relationship
    GetRelationshipsForEntity(entityID string) []*Relationship
    GetTopologicalOrder() ([]string, error)
}

// EntityInterface defines the operations that can be performed on an Entity
type EntityInterface interface {
    GetID() string
    GetExternalID() string
    GetDisplayName() string
    GetDescription() string
    GetAttributes() []*Attribute
    GetPrimaryKey() *Attribute
    // Include any other methods from your Entity struct
}

// RelationshipInterface defines operations for relationships
type RelationshipInterface interface {
    GetID() string
    GetSourceEntity() EntityInterface
    GetTargetEntity() EntityInterface
    GetSourceAttribute() *Attribute
    GetTargetAttribute() *Attribute
    // Include any other methods from your Relationship struct
}

// AttributeInterface defines operations for attributes
type AttributeInterface interface {
    GetName() string
    GetExternalID() string
    GetType() string
    IsUnique() bool
    GetDescription() string
    GetParentEntity() EntityInterface
    // Include any other methods from your Attribute struct
}
```

## Step 2: Update Your Implementations to Use the Interfaces

Ensure your concrete types implement these interfaces:

```go
// Ensure Graph implements GraphInterface
var _ GraphInterface = (*Graph)(nil)

// Ensure Entity implements EntityInterface
var _ EntityInterface = (*Entity)(nil)

// Ensure Relationship implements RelationshipInterface
var _ RelationshipInterface = (*Relationship)(nil)

// Ensure Attribute implements AttributeInterface
var _ AttributeInterface = (*Attribute)(nil)
```

## Step 3: Update Constructors to Return Interfaces

Modify your constructors to return interfaces rather than concrete types:

```go
// NewGraph creates a new Graph from the YAML model
func NewGraph(yamlModel *models.SORDefinition) (GraphInterface, error) {
    // Existing implementation remains the same
    // ...

    return graph, nil
}

// Other constructors should be updated similarly
func newEntity(...) (EntityInterface, error) {
    // ...
}

func newRelationship(...) (RelationshipInterface, error) {
    // ...
}
```

## Step 4: Improve Dependency Injection

Consider making your functions accept interfaces instead of concrete types:

```go
// Before:
func ProcessGraph(g *Graph) {
    // ...
}

// After:
func ProcessGraph(g GraphInterface) {
    // ...
}
```

## Step 5: Example Test with Testify Mock

Here's how you would test a function that uses your Graph with testify mocks:

```go
package model_test

import (
    "testing"
    
    "github.com/SGNL-ai/fabricator/pkg/model" // Import your package
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockGraph is a testify mock implementation of GraphInterface
type MockGraph struct {
    mock.Mock
}

// Implement all required GraphInterface methods
func (m *MockGraph) GetEntity(id string) (*model.Entity, bool) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Bool(1)
    }
    return args.Get(0).(*model.Entity), args.Bool(1)
}

func (m *MockGraph) GetAllEntities() map[string]*model.Entity {
    args := m.Called()
    return args.Get(0).(map[string]*model.Entity)
}

func (m *MockGraph) GetEntitiesList() []*model.Entity {
    args := m.Called()
    return args.Get(0).([]*model.Entity)
}

func (m *MockGraph) GetRelationship(id string) (*model.Relationship, bool) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Bool(1)
    }
    return args.Get(0).(*model.Relationship), args.Bool(1)
}

func (m *MockGraph) GetAllRelationships() []*model.Relationship {
    args := m.Called()
    return args.Get(0).([]*model.Relationship)
}

func (m *MockGraph) GetRelationshipsForEntity(entityID string) []*model.Relationship {
    args := m.Called(entityID)
    return args.Get(0).([]*model.Relationship)
}

func (m *MockGraph) GetTopologicalOrder() ([]string, error) {
    args := m.Called()
    return args.Get(0).([]string), args.Error(1)
}

// Similar mocks should be created for EntityInterface, RelationshipInterface, etc.

// Example test
func TestGraphProcessor(t *testing.T) {
    // Create mock graph
    mockGraph := new(MockGraph)
    
    // Set up expectations
    mockEntities := make(map[string]*model.Entity)
    mockEntities["entity1"] = &model.Entity{} // Initialize with test data
    
    mockGraph.On("GetAllEntities").Return(mockEntities)
    mockGraph.On("GetRelationshipsForEntity", "entity1").Return([]*model.Relationship{})
    
    // Call function under test
    result := model.ProcessGraph(mockGraph)
    
    // Assert expectations and results
    mockGraph.AssertExpectations(t)
    assert.Equal(t, expectedResult, result)
}
```

## Step 6: Internal Methods and Testing

For internal methods like `createEntitiesFromYAML`, you have a few options:

1. **Make them exportable** (capitalize first letter) if testing them directly is valuable
2. **Extract complex logic** to separate testable types
3. **Test them indirectly** through the public API

Example of option 2 (extraction):

```go
// EntityBuilder handles entity creation logic
type EntityBuilder struct {
    // ...
}

func (b *EntityBuilder) CreateEntitiesFromYAML(yamlEntities map[string]models.Entity) (map[string]EntityInterface, error) {
    // Logic moved from Graph.createEntitiesFromYAML
}

// Graph would then use this builder
func NewGraph(yamlModel *models.SORDefinition) (GraphInterface, error) {
    // ...
    builder := NewEntityBuilder()
    entities, err := builder.CreateEntitiesFromYAML(yamlModel.Entities)
    // ...
}
```

## Conclusion

By implementing these changes, your Graph model will be much easier to test with mocks. The interfaces provide a clear contract for what each component does, and the testify mocks allow you to verify that your code interacts with the Graph correctly.

Remember that good mocking is about testing the interactions between components, not the internal implementation details of those components. The mocks allow you to test that a function uses the Graph as expected, without requiring a real Graph implementation.
