package reader

import (
	"context"
	"fmt"
	"os"

	"github.com/leocomelli/aigile/internal/prompt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsService is an interface for the minimal Google Sheets API used by the reader.
type SheetsService interface {
	GetValues(spreadsheetID, readRange string) ([][]interface{}, error)
}

// realSheetsService implements SheetsService using the real Google Sheets API.
type realSheetsService struct {
	srv *sheets.Service
}

func (r *realSheetsService) GetValues(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := r.srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

// GoogleSheetsReader reads items from a Google Sheets spreadsheet.
type GoogleSheetsReader struct {
	SpreadsheetID   string
	CredentialsFile string        // Caminho para o arquivo de credenciais JSON
	SheetsAPI       SheetsService // opcional, para testes
}

// DefaultGoogleSheetRange is the default range read from Google Sheets.
const DefaultGoogleSheetRange = "Sheet1!A:D"

// NewGoogleSheetsReader creates a new reader for Google Sheets.
func NewGoogleSheetsReader(spreadsheetID, credentialsFile string) *GoogleSheetsReader {
	return &GoogleSheetsReader{
		SpreadsheetID:   spreadsheetID,
		CredentialsFile: credentialsFile,
	}
}

// NewGoogleSheetsReaderWithService allows injecting a custom SheetsService (para testes).
func NewGoogleSheetsReaderWithService(spreadsheetID, credentialsFile string, service SheetsService) *GoogleSheetsReader {
	return &GoogleSheetsReader{
		SpreadsheetID:   spreadsheetID,
		CredentialsFile: credentialsFile,
		SheetsAPI:       service,
	}
}

func (r *GoogleSheetsReader) Read() ([]Item, error) {
	var service SheetsService
	if r.SheetsAPI != nil {
		service = r.SheetsAPI
	} else {
		ctx := context.Background()
		b, err := os.ReadFile(r.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read credentials file: %w", err)
		}
		config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsReadonlyScope)
		if err != nil {
			return nil, fmt.Errorf("unable to parse credentials file: %w", err)
		}
		client := config.Client(ctx)
		srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve Sheets client: %w", err)
		}
		service = &realSheetsService{srv: srv}
	}

	respValues, err := service.GetValues(r.SpreadsheetID, DefaultGoogleSheetRange)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	var items []Item
	for i, row := range respValues {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 4 {
			continue
		}
		itemType := prompt.ItemType(fmt.Sprintf("%v", row[0]))
		item := Item{
			Type:    itemType,
			Parent:  fmt.Sprintf("%v", row[1]),
			Context: fmt.Sprintf("%v", row[2]),
		}
		if len(row) > 3 {
			for _, c := range row[3:] {
				item.Criteria = append(item.Criteria, fmt.Sprintf("%v", c))
			}
		}
		items = append(items, item)
	}
	return items, nil
}
