package pipeline

import (
	"errors"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type MockIDGenerator struct {
	mock.Mock
}

// Ensure MockIDGenerator implements the IDGeneratorInterface
var _ IDGeneratorInterface = (*MockIDGenerator)(nil)

func (m *MockIDGenerator) GenerateIDs(graph *model.Graph, rowCounts map[string]int) error {
	args := m.Called(graph, rowCounts)
	return args.Error(0)
}

type MockRelationshipLinker struct {
	mock.Mock
}

// Ensure MockRelationshipLinker implements the RelationshipLinkerInterface
var _ RelationshipLinkerInterface = (*MockRelationshipLinker)(nil)

func (m *MockRelationshipLinker) LinkRelationships(graph *model.Graph, autoCardinality bool) error {
	args := m.Called(graph, autoCardinality)
	return args.Error(0)
}

type MockFieldGenerator struct {
	mock.Mock
}

// Ensure MockFieldGenerator implements the FieldGeneratorInterface
var _ FieldGeneratorInterface = (*MockFieldGenerator)(nil)

func (m *MockFieldGenerator) GenerateFields(graph *model.Graph) error {
	args := m.Called(graph)
	return args.Error(0)
}

type MockValidator struct {
	mock.Mock
}

// Ensure MockValidator implements the ValidatorInterface
var _ ValidatorInterface = (*MockValidator)(nil)

func (m *MockValidator) ValidateRelationships(graph *model.Graph) []string {
	args := m.Called(graph)
	return args.Get(0).([]string)
}

func (m *MockValidator) ValidateUniqueValues(graph *model.Graph) []string {
	args := m.Called(graph)
	return args.Get(0).([]string)
}

type MockCSVWriter struct {
	mock.Mock
}

// Ensure MockCSVWriter implements the CSVWriterInterface
var _ CSVWriterInterface = (*MockCSVWriter)(nil)

func (m *MockCSVWriter) WriteFiles(graph *model.Graph) error {
	args := m.Called(graph)
	return args.Error(0)
}

func TestNewDataGenerator(t *testing.T) {
	tests := []struct {
		name            string
		outputDir       string
		rowCounts       map[string]int
		autoCardinality bool
	}{
		{
			name:            "should create generator with valid parameters",
			outputDir:       "/tmp/output",
			rowCounts:       map[string]int{"entity1": 100, "entity2": 200},
			autoCardinality: true,
		},
		{
			name:            "should create generator with default parameters",
			outputDir:       "output",
			rowCounts:       map[string]int{"entity1": 10},
			autoCardinality: false,
		},
		{
			name:            "should handle empty output directory",
			outputDir:       "",
			rowCounts:       map[string]int{"entity1": 50},
			autoCardinality: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewDataGenerator(tt.outputDir, tt.rowCounts, tt.autoCardinality)

			require.NotNil(t, generator)
			assert.Equal(t, tt.outputDir, generator.outputDir)
			assert.Equal(t, tt.rowCounts, generator.rowCounts)
			assert.Equal(t, tt.autoCardinality, generator.autoCardinality)

			// Verify all components are initialized
			assert.NotNil(t, generator.idGenerator)
			assert.NotNil(t, generator.relationshipLinker)
			assert.NotNil(t, generator.fieldGenerator)
			assert.NotNil(t, generator.validator)
			assert.NotNil(t, generator.csvWriter)

			// Verify the components are of the correct types
			assert.IsType(t, &IDGenerator{}, generator.idGenerator)
			assert.IsType(t, &RelationshipLinker{}, generator.relationshipLinker)
			assert.IsType(t, &FieldGenerator{}, generator.fieldGenerator)
			assert.IsType(t, &Validator{}, generator.validator)
			assert.IsType(t, &CSVWriter{}, generator.csvWriter)
		})
	}
}

func TestDataGenerator_Generate(t *testing.T) {
	tests := []struct {
		name            string
		setupMocks      func(*MockIDGenerator, *MockRelationshipLinker, *MockFieldGenerator, *MockValidator, *MockCSVWriter)
		graph           *model.Graph
		dataVolume      int
		autoCardinality bool
		wantErr         bool
		expectedError   string
	}{
		{
			name: "Successful generation",
			setupMocks: func(idGen *MockIDGenerator, relLinker *MockRelationshipLinker, fieldGen *MockFieldGenerator, validator *MockValidator, csvWriter *MockCSVWriter) {
				// All phases succeed
				idGen.On("GenerateIDs", mock.Anything, mock.Anything).Return(nil)
				relLinker.On("LinkRelationships", mock.Anything, false).Return(nil)
				fieldGen.On("GenerateFields", mock.Anything).Return(nil)
				csvWriter.On("WriteFiles", mock.Anything).Return(nil)
			},
			graph:           nil, // Will be initialized in test
			dataVolume:      10,
			autoCardinality: false,
			wantErr:         false,
		},
		{
			name: "ID generation fails",
			setupMocks: func(idGen *MockIDGenerator, relLinker *MockRelationshipLinker, fieldGen *MockFieldGenerator, validator *MockValidator, csvWriter *MockCSVWriter) {
				// ID generation fails
				idGen.On("GenerateIDs", mock.Anything, mock.Anything).Return(errors.New("ID generation error"))
				// Other mocks shouldn't be called
			},
			graph:           nil, // Will be initialized in test
			dataVolume:      10,
			autoCardinality: false,
			wantErr:         true,
			expectedError:   "ID generation failed: ID generation error",
		},
		{
			name: "Relationship linking fails",
			setupMocks: func(idGen *MockIDGenerator, relLinker *MockRelationshipLinker, fieldGen *MockFieldGenerator, validator *MockValidator, csvWriter *MockCSVWriter) {
				// ID generation succeeds
				idGen.On("GenerateIDs", mock.Anything, mock.Anything).Return(nil)
				// Relationship linking fails
				relLinker.On("LinkRelationships", mock.Anything, false).Return(errors.New("relationship error"))
				// Other mocks shouldn't be called
			},
			graph:           nil, // Will be initialized in test
			dataVolume:      10,
			autoCardinality: false,
			wantErr:         true,
			expectedError:   "relationship linking failed: relationship error",
		},
		{
			name: "Field generation fails",
			setupMocks: func(idGen *MockIDGenerator, relLinker *MockRelationshipLinker, fieldGen *MockFieldGenerator, validator *MockValidator, csvWriter *MockCSVWriter) {
				// ID generation succeeds
				idGen.On("GenerateIDs", mock.Anything, mock.Anything).Return(nil)
				// Relationship linking succeeds
				relLinker.On("LinkRelationships", mock.Anything, false).Return(nil)
				// Field generation fails
				fieldGen.On("GenerateFields", mock.Anything).Return(errors.New("field generation error"))
				// Other mocks shouldn't be called after field generation fails
			},
			graph:           nil,
			dataVolume:      10,
			autoCardinality: false,
			wantErr:         true,
			expectedError:   "field generation failed: field generation error",
		},
		{
			name: "CSV writing fails",
			setupMocks: func(idGen *MockIDGenerator, relLinker *MockRelationshipLinker, fieldGen *MockFieldGenerator, validator *MockValidator, csvWriter *MockCSVWriter) {
				// All steps succeed until CSV writing
				idGen.On("GenerateIDs", mock.Anything, mock.Anything).Return(nil)
				relLinker.On("LinkRelationships", mock.Anything, false).Return(nil)
				fieldGen.On("GenerateFields", mock.Anything).Return(nil)
				// CSV writing fails
				csvWriter.On("WriteFiles", mock.Anything).Return(errors.New("CSV writing error"))
			},
			graph:           nil,
			dataVolume:      10,
			autoCardinality: false,
			wantErr:         true,
			expectedError:   "CSV file writing failed: CSV writing error",
		},
		{
			name: "Validation errors are logged but don't fail generation",
			setupMocks: func(idGen *MockIDGenerator, relLinker *MockRelationshipLinker, fieldGen *MockFieldGenerator, validator *MockValidator, csvWriter *MockCSVWriter) {
				// All phases succeed
				idGen.On("GenerateIDs", mock.Anything, mock.Anything).Return(nil)
				relLinker.On("LinkRelationships", mock.Anything, false).Return(nil)
				fieldGen.On("GenerateFields", mock.Anything).Return(nil)
				// No validation expectations since validation was removed from generation
				csvWriter.On("WriteFiles", mock.Anything).Return(nil)
			},
			graph:           nil,
			dataVolume:      10,
			autoCardinality: false,
			wantErr:         false, // Validation errors don't cause failure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock instances
			mockIDGen := new(MockIDGenerator)
			mockRelLinker := new(MockRelationshipLinker)
			mockFieldGen := new(MockFieldGenerator)
			mockValidator := new(MockValidator)
			mockCSVWriter := new(MockCSVWriter)

			// Setup expectations
			tt.setupMocks(mockIDGen, mockRelLinker, mockFieldGen, mockValidator, mockCSVWriter)

			// Create generator with mocks
			generator := &DataGenerator{
				idGenerator:        mockIDGen,
				relationshipLinker: mockRelLinker,
				fieldGenerator:     mockFieldGen,
				validator:          mockValidator,
				csvWriter:          mockCSVWriter,
				rowCounts:          map[string]int{"entity1": tt.dataVolume},
				autoCardinality:    tt.autoCardinality,
			}

			// Initialize a graph if needed
			if tt.graph == nil {
				// In a real test we would create a proper graph
				tt.graph = &model.Graph{}
			}

			// Execute the generator
			err := generator.Generate(tt.graph)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify that mocks were called as expected
			mockIDGen.AssertExpectations(t)
			mockRelLinker.AssertExpectations(t)
			mockFieldGen.AssertExpectations(t)
			mockValidator.AssertExpectations(t)
			mockCSVWriter.AssertExpectations(t)
		})
	}
}
