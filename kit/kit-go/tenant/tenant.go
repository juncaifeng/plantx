package tenant

// Info holds the resolved tenant for the current request.
type Info struct {
	ID   string
	Name string
}

// Resolver derives tenant information from an authenticated user.
type Resolver interface {
	Resolve(userID string, claims map[string]string) (Info, error)
}
