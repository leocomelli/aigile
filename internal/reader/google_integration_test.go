//go:build integration || integration_test
// +build integration integration_test

package reader

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleSheetsReader_Integration_Read(t *testing.T) {
	spreadsheetID := os.Getenv("GOOGLE_SHEET_ID")
	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")

	if spreadsheetID == "" || credentialsFile == "" {
		t.Skip("GOOGLE_SHEET_ID and GOOGLE_CREDENTIALS_FILE must be set for integration test")
	}

	r := NewGoogleSheetsReader(spreadsheetID, credentialsFile)
	items, err := r.Read()
	require.NoError(t, err, "Failed to read from Google Sheets. Check if the ID and credentials are correct.")
	require.NotNil(t, items)
	t.Logf("Items read: %+v", items)

	if len(items) > 0 {
		item := items[0]
		assert.NotEmpty(t, item.Type)
		assert.NotEmpty(t, item.Parent)
		assert.NotEmpty(t, item.Context)
	}
}

// Integration tests for GoogleSheetsReader should be placed in a separate file with build tag 'integration'.
