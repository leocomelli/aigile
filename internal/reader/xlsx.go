// Package reader provides utilities to read and parse XLSX files for aigile.
package reader

import (
	"fmt"

	"log/slog"

	"github.com/leocomelli/aigile/internal/prompt"
	"github.com/xuri/excelize/v2"
)

// Item represents a row read from a source (XLSX, Google Sheets, etc).
type Item = struct {
	Type     prompt.ItemType
	Parent   string
	Context  string
	Criteria []string
}

// XLSXReader reads items from an XLSX file.
type XLSXReader struct {
	filePath string
}

// NewXLSXReader creates a new XLSXReader for the given file path.
func NewXLSXReader(filePath string) *XLSXReader {
	return &XLSXReader{
		filePath: filePath,
	}
}

// Read reads the XLSX file and returns a slice of Items or an error.
func (r *XLSXReader) Read() ([]Item, error) {
	f, err := excelize.OpenFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Warn("failed to close xlsx file", "error", err)
		}
	}()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("failed to get rows: no sheets found")
	}

	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("failed to get rows: sheet '%s' is empty or invalid", sheetName)
	}

	var items []Item
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 4 {
			continue
		}

		// Convert string type to ItemType
		itemType := prompt.ItemType(row[0])
		if !itemType.IsValid() {
			return nil, fmt.Errorf("invalid item type at row %d: %s", i+1, row[0])
		}

		item := Item{
			Type:    itemType,
			Parent:  row[1],
			Context: row[2],
		}

		// Add criteria if available
		if len(row) > 3 {
			item.Criteria = row[3:]
		}

		items = append(items, item)
	}

	return items, nil
}
