package domain

// User represents a platform user.
type User struct {
	ID       string
	Username string
	TenantID string
	RoleIDs  []string
}

// Role represents a user role.
type Role struct {
	ID          string
	Name        string
	Permissions []string
	Description string
}

// Permission represents a granular authorization permission.
type Permission struct {
	ID          string
	Name        string
	Resource    string
	Operation   string
	Description string
}
