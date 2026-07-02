package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIAMClientSyncPermissions(t *testing.T) {
	var created []Permission
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/iam/v1/permissions":
			switch r.Method {
			case http.MethodGet:
				_ = json.NewEncoder(w).Encode(iamPermissionsResponse{
					Permissions: []iamPermission{
						{ID: "1", Name: "item:list", Resource: "item", Operation: "list"},
					},
				})
			case http.MethodPost:
				var req iamCreatePermissionRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				created = append(created, Permission(req))
				w.WriteHeader(http.StatusCreated)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewIAMClient(server.URL)
	declared := []Permission{
		{Name: "item:list", Resource: "item", Operation: "list"},
		{Name: "setting:admin", Resource: "setting", Operation: "admin"},
	}
	if err := client.SyncPermissions(context.Background(), declared); err != nil {
		t.Fatalf("sync permissions: %v", err)
	}

	if len(created) != 1 || created[0].Name != "setting:admin" {
		t.Errorf("created permissions = %+v, want [setting:admin]", created)
	}
}

func TestIAMClientValidatePermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/iam/v1/permissions" && r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(iamPermissionsResponse{
				Permissions: []iamPermission{
					{ID: "1", Name: "item:list", Resource: "item", Operation: "list"},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewIAMClient(server.URL)
	declared := []Permission{{Name: "setting:admin", Resource: "setting", Operation: "admin"}}
	required := []string{"item:list", "setting:admin"}
	if err := client.ValidatePermissions(context.Background(), declared, required); err != nil {
		t.Fatalf("validate permissions: %v", err)
	}

	if err := client.ValidatePermissions(context.Background(), declared, []string{"unknown:perm"}); err == nil {
		t.Error("expected validation error for unknown permission")
	}
}
