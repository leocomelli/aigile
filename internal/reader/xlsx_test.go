package reader

import (
	"fmt"
	"os"
	"testing"

	"github.com/leocomelli/aigile/internal/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// createTestXLSX creates a temporary XLSX file with the provided rows for testing.
func createTestXLSX(t *testing.T, rows [][]string) string {
	f := excelize.NewFile()
	// Rename default sheet to 'Sheet1' if needed
	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())
	if defaultSheet != "Sheet1" {
		if err := f.SetSheetName(defaultSheet, "Sheet1"); err != nil {
			t.Fatalf("failed to set sheet name: %v", err)
		}
	}
	// Remove all sheets except 'Sheet1'
	for _, name := range f.GetSheetList() {
		if name != "Sheet1" {
			if err := f.DeleteSheet(name); err != nil {
				t.Fatalf("failed to delete sheet: %v", err)
			}
		}
	}
	// Write data to 'Sheet1'
	for i, row := range rows {
		rowNum := i + 1
		for j, cell := range row {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cellName := fmt.Sprintf("%s%d", col, rowNum)
			if err := f.SetCellValue("Sheet1", cellName, cell); err != nil {
				t.Fatalf("failed to set cell value: %v", err)
			}
		}
	}
	idx, _ := f.GetSheetIndex("Sheet1")
	f.SetActiveSheet(idx)
	file, err := os.CreateTemp("", "test-*.xlsx")
	assert.NoError(t, err)
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("failed to close file: %v", err)
		}
	}()
	assert.NoError(t, f.SaveAs(file.Name()))
	return file.Name()
}

// TestXLSXReader_Read_Success tests reading a valid XLSX file and parsing items successfully.
func TestXLSXReader_Read_Success(t *testing.T) {
	rows := [][]string{
		{"Type", "Parent", "Context", "Criteria1", "Criteria2"},
		{"User Story", "FEAT-1", "Context1", "Crit1", "Crit2"},
	}
	file := createTestXLSX(t, rows)
	defer func() {
		if err := os.Remove(file); err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}
	}()

	r := NewXLSXReader(file)
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, prompt.UserStory, items[0].Type)
	assert.Equal(t, "FEAT-1", items[0].Parent)
	assert.Equal(t, "Context1", items[0].Context)
	assert.Equal(t, []string{"Crit1", "Crit2"}, items[0].Criteria)
}

// TestXLSXReader_Read_OpenFileError tests error handling when the XLSX file does not exist.
func TestXLSXReader_Read_OpenFileError(t *testing.T) {
	r := NewXLSXReader("nonexistent.xlsx")
	items, err := r.Read()
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to open file")
}

// TestXLSXReader_Read_GetRowsError tests error handling when the sheet is missing or invalid.
func TestXLSXReader_Read_GetRowsError(t *testing.T) {
	// Cria um arquivo válido, mas remove a planilha 'Sheet1' para simular erro
	f := excelize.NewFile()
	if err := f.DeleteSheet("Sheet1"); err != nil {
		t.Fatalf("failed to delete sheet: %v", err)
	}
	file, err := os.CreateTemp("", "test-*.xlsx")
	assert.NoError(t, err)
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("failed to close file: %v", err)
		}
	}()
	assert.NoError(t, f.SaveAs(file.Name()))

	r := NewXLSXReader(file.Name())
	items, err := r.Read()
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to get rows")
}

// TestXLSXReader_Read_InvalidType tests error handling for invalid item types in the XLSX file.
func TestXLSXReader_Read_InvalidType(t *testing.T) {
	rows := [][]string{
		{"Type", "Parent", "Context", "Criteria1"},
		{"InvalidType", "FEAT-1", "Context1", "Crit1"},
	}
	file := createTestXLSX(t, rows)
	defer func() {
		if err := os.Remove(file); err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}
	}()

	r := NewXLSXReader(file)
	items, err := r.Read()
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "invalid item type")
}

// TestXLSXReader_Read_SkipHeaderAndShortRows tests skipping header and short/incomplete rows.
func TestXLSXReader_Read_SkipHeaderAndShortRows(t *testing.T) {
	rows := [][]string{
		{"Type", "Parent", "Context", "Criteria1"},
		{"User Story", "FEAT-1", "Context1"},          // too short
		{"User Story", "FEAT-2", "Context2", "Crit1"}, // valid
	}
	file := createTestXLSX(t, rows)
	defer func() {
		if err := os.Remove(file); err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}
	}()

	r := NewXLSXReader(file)
	items, err := r.Read()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "FEAT-2", items[0].Parent)
}
