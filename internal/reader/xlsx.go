package reader

import (
	"fmt"

	"github.com/leocomelli/aigile/internal/prompt"
	"github.com/xuri/excelize/v2"
)

type Item struct {
	Type     prompt.ItemType
	Parent   string
	Context  string
	Criteria []string
}

type XLSXReader struct {
	filePath string
}

func NewXLSXReader(filePath string) *XLSXReader {
	return &XLSXReader{
		filePath: filePath,
	}
}

func (r *XLSXReader) Read() ([]Item, error) {
	f, err := excelize.OpenFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
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
