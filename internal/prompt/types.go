package prompt

// ItemType represents the type of agile item
type ItemType string

const (
	Epic      ItemType = "Epic"
	Feature   ItemType = "Feature"
	UserStory ItemType = "User Story"
	Task      ItemType = "Task"
)

// IsValid checks if the item type is valid
func (t ItemType) IsValid() bool {
	switch t {
	case Epic, Feature, UserStory, Task:
		return true
	default:
		return false
	}
}

// String returns the string representation of the item type
func (t ItemType) String() string {
	return string(t)
}
