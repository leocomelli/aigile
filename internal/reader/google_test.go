package reader

import (
	"errors"
	"os"
	"testing"

	"github.com/leocomelli/aigile/internal/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---
type mockSheetsService struct {
	values [][]interface{}
	err    error
}

func (m *mockSheetsService) GetValues(spreadsheetID, readRange string) ([][]interface{}, error) {
	return m.values, m.err
}

// --- Unit tests ---

func TestGoogleSheetsReader_Read_InvalidCredentialsFile(t *testing.T) {
	t.Parallel()
	r := NewGoogleSheetsReader("spreadsheet-id", "nonexistent.json")
	items, err := r.Read()
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "unable to read credentials file")
}

func TestGoogleSheetsReader_Read_InvalidCredentialsContent(t *testing.T) {
	t.Parallel()
	file, err := os.CreateTemp("", "invalid-creds-*.json")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString("not a json")
	require.NoError(t, err)
	file.Close()

	r := NewGoogleSheetsReader("spreadsheet-id", file.Name())
	items, err := r.Read()
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "unable to parse credentials file")
}

func TestGoogleSheetsReader_Read_EmptySheet(t *testing.T) {
	r := NewGoogleSheetsReaderWithService("id", "creds", &mockSheetsService{values: [][]interface{}{}})
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestGoogleSheetsReader_Read_HeaderOnly(t *testing.T) {
	r := NewGoogleSheetsReaderWithService("id", "creds", &mockSheetsService{values: [][]interface{}{{"Type", "Parent", "Context", "Criteria"}}})
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestGoogleSheetsReader_Read_IncompleteRow(t *testing.T) {
	values := [][]interface{}{
		{"Type", "Parent", "Context", "Criteria"},
		{"User Story", "Parent1", "Context1"}, // incomplete
	}
	r := NewGoogleSheetsReaderWithService("id", "creds", &mockSheetsService{values: values})
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestGoogleSheetsReader_Read_InvalidType(t *testing.T) {
	values := [][]interface{}{
		{"Type", "Parent", "Context", "Criteria"},
		{"InvalidType", "Parent1", "Context1", "Crit1"},
	}
	r := NewGoogleSheetsReaderWithService("id", "creds", &mockSheetsService{values: values})
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, prompt.ItemType("InvalidType"), items[0].Type)
}

func TestGoogleSheetsReader_Read_ValidRow(t *testing.T) {
	values := [][]interface{}{
		{"Type", "Parent", "Context", "Criteria1", "Criteria2"},
		{"User Story", "FEAT-1", "Context1", "Crit1", "Crit2"},
	}
	r := NewGoogleSheetsReaderWithService("id", "creds", &mockSheetsService{values: values})
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, prompt.UserStory, items[0].Type)
	assert.Equal(t, "FEAT-1", items[0].Parent)
	assert.Equal(t, "Context1", items[0].Context)
	assert.Equal(t, []string{"Crit1", "Crit2"}, items[0].Criteria)
}

func TestGoogleSheetsReader_Read_ServiceError(t *testing.T) {
	r := NewGoogleSheetsReaderWithService("id", "creds", &mockSheetsService{err: errors.New("fail")})
	items, err := r.Read()
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "unable to retrieve data from sheet")
}
