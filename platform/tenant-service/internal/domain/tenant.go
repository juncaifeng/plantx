package domain

// Tenant represents a platform tenant.
type Tenant struct {
	ID        string
	Name      string
	Status    string
	CreatedAt int64
}
