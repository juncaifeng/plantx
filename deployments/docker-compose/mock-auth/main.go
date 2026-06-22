package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	signingKey *rsa.PrivateKey
	kid        string
)

func keyPath() string {
	if p := os.Getenv("KEY_PATH"); p != "" {
		return p
	}
	return "/data/mock-auth-key.pem"
}

func loadOrGenerateKey() (*rsa.PrivateKey, error) {
	path := keyPath()
	data, err := os.ReadFile(path)
	if err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				return key, nil
			}
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		return nil, err
	}
	return key, nil
}

func init() {
	var err error
	signingKey, err = loadOrGenerateKey()
	if err != nil {
		panic(err)
	}
	h := sha256.New()
	h.Write(signingKey.PublicKey.N.Bytes())
	kid = base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:16]
}

type jwk struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

func jwksHandler(w http.ResponseWriter, r *http.Request) {
	resp := jwksResponse{
		Keys: []jwk{{
			Kty: "RSA",
			Use: "sig",
			Kid: kid,
			N:   base64.RawURLEncoding.EncodeToString(signingKey.PublicKey.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(signingKey.PublicKey.E)).Bytes()),
		}},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		username = "demo"
	}

	now := time.Now()
	var tenantID string
	var roles []string
	var permissions []string

	switch username {
	case "admin":
		tenantID = "platform"
		roles = []string{"admin"}
		permissions = []string{
			"platform:admin",
			"tenant:create", "tenant:read", "tenant:list",
			"iam:create", "iam:read", "iam:list",
			"gateway:create", "gateway:read", "gateway:list",
			"audit:read", "audit:list",
			"service:list", "service:read", "service:delete",
			"microapp:create", "microapp:list", "microapp:update", "microapp:delete",
			"menu:create", "menu:list", "menu:update", "menu:delete",
			"permission:create", "permission:list", "permission:delete",
			"role:create", "role:read", "role:update", "role:delete",
			"route:read", "route:update", "route:sync",
		}
	default:
		tenantID = "t_001"
		if strings.HasSuffix(username, "-b") {
			tenantID = "t_002"
		}
		roles = []string{"user"}
		permissions = []string{"order:create", "order:read", "order:list"}
	}

	claims := jwt.MapClaims{
		"iss":         "http://mock-auth:8080",
		"sub":         username,
		"aud":         "plantx-portal",
		"exp":         now.Add(time.Hour).Unix(),
		"iat":         now.Unix(),
		"tenant_id":   tenantID,
		"roles":       roles,
		"permissions": permissions,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(signingKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": signed,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func main() {
	http.HandleFunc("/.well-known/jwks.json", jwksHandler)
	http.HandleFunc("/oauth/token", tokenHandler)
	fmt.Println("mock auth server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
