package prompt

// ItemType represents the type of agile item
type ItemType string

// UserStory represents the 'User Story' agile item type.
const (
	UserStory ItemType = "User Story"
)

// IsValid checks if the item type is valid
func (t ItemType) IsValid() bool {
	switch t {
	case UserStory:
		return true
	default:
		return false
	}
}

// String returns the string representation of the item type
func (t ItemType) String() string {
	return string(t)
}
