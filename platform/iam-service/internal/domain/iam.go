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

// Attribute defines an ABAC attribute key and its value type.
type Attribute struct {
	ID          string
	Key         string
	ValueType   string
	Description string
}

// Condition is a single ABAC comparison.
type Condition struct {
	ID           string
	Name         string
	AttributeKey string
	Operator     string
	Value        string
	Description  string
}

// Policy combines permissions with conditions for ABAC decisions.
type Policy struct {
	ID          string
	Name        string
	Description string
	Permissions []string
	ConditionIDs []string
	Effect      string
	Priority    int32
}
