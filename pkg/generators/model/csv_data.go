package model

// CSVData represents a structure to hold data for CSV file generation
type CSVData struct {
	ExternalId  string
	Headers     []string
	Rows        [][]string
	EntityName  string
	Description string
}
