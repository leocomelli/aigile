package reader

// Reader is the interface for reading items from a source (XLSX, Google Sheets, etc).
type Reader interface {
	Read() ([]Item, error)
}
