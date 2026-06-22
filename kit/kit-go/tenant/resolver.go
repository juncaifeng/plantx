// Package tenant provides tenant resolution abstractions and implementations.
package tenant

// FromUser resolves tenant info from user claims.
type FromUser struct{}

// NewResolver creates a tenant resolver that extracts tenant_id from claims.
func NewResolver() *FromUser {
	return &FromUser{}
}

// Resolve extracts tenant info from the user ID and claims.
func (r *FromUser) Resolve(_ string, claims map[string]string) (Info, error) {
	tenantID := claims["tenant_id"]
	if tenantID == "" {
		tenantID = claims["org_id"]
	}
	name := claims["tenant_name"]
	if name == "" {
		name = tenantID
	}
	return Info{ID: tenantID, Name: name}, nil
}
