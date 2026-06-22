package maxkey

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/plantx/kit/kit-go/auth"
)

// Options configures the MaxKey authenticator.
type Options struct {
	Issuer       string
	JWKSURL      string
	PublicKeyPEM string
	// SkipVerify is for testing/stub mode only.
	SkipVerify bool
}

// Authenticator validates OIDC/JWT tokens issued by MaxKey.
type Authenticator struct {
	opts   Options
	mu     sync.RWMutex
	keys   map[string]*rsa.PublicKey
	client *http.Client
}

// New creates a MaxKey authenticator.
func New(opts Options) *Authenticator {
	return &Authenticator{
		opts:   opts,
		keys:   make(map[string]*rsa.PublicKey),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Authenticate validates a bearer token and returns user info.
func (a *Authenticator) Authenticate(_ context.Context, credential string) (*auth.UserInfo, error) {
	token := strings.TrimPrefix(credential, "Bearer ")
	if token == "" {
		return nil, fmt.Errorf("missing bearer token")
	}

	// Stub mode: accept well-formed JWTs without signature verification.
	if a.opts.SkipVerify {
		return a.parseUnverified(token)
	}

	parser := jwt.NewParser(jwt.WithIssuer(a.opts.Issuer))
	parsed, err := parser.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		return a.getKey(kid)
	})
	if err != nil {
		return nil, err
	}
	return claimsToUser(parsed.Claims.(jwt.MapClaims))
}

func (a *Authenticator) parseUnverified(token string) (*auth.UserInfo, error) {
	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	return claimsToUser(parsed.Claims.(jwt.MapClaims))
}

func claimsToUser(c jwt.MapClaims) (*auth.UserInfo, error) {
	sub, _ := c["sub"].(string)
	if sub == "" {
		sub, _ = c["id"].(string)
	}
	if sub == "" {
		return nil, fmt.Errorf("token missing subject claim")
	}
	tenantID, _ := c["tenant_id"].(string)
	if tenantID == "" {
		tenantID, _ = c["org_id"].(string)
	}
	username, _ := c["preferred_username"].(string)
	if username == "" {
		username, _ = c["username"].(string)
	}
	roles := stringSlice(c["roles"])
	perms := stringSlice(c["permissions"])
	claims := map[string]string{}
	for k, v := range c {
		if s, ok := v.(string); ok {
			claims[k] = s
		}
	}
	return &auth.UserInfo{
		ID:          sub,
		TenantID:    tenantID,
		Username:    username,
		DisplayName: stringClaim(c, "name"),
		Email:       stringClaim(c, "email"),
		Roles:       roles,
		Permissions: perms,
		Claims:      claims,
	}, nil
}

func stringClaim(c jwt.MapClaims, key string) string {
	v, _ := c[key].(string)
	return v
}

func stringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, e := range val {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		if val == "" {
			return nil
		}
		return []string{val}
	}
	return nil
}

func (a *Authenticator) getKey(kid string) (interface{}, error) {
	a.mu.RLock()
	key := a.keys[kid]
	a.mu.RUnlock()
	if key != nil {
		return key, nil
	}
	if a.opts.PublicKeyPEM != "" {
		pk, err := parsePublicKeyPEM(a.opts.PublicKeyPEM)
		if err != nil {
			return nil, err
		}
		a.mu.Lock()
		a.keys[kid] = pk
		a.mu.Unlock()
		return pk, nil
	}
	if a.opts.JWKSURL != "" {
		if err := a.refreshKeys(); err != nil {
			return nil, err
		}
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	if key := a.keys[kid]; key != nil {
		return key, nil
	}
	return nil, fmt.Errorf("unknown key id: %s", kid)
}

type jwks struct {
	Keys []map[string]any `json:"keys"`
}

func (a *Authenticator) refreshKeys() error {
	resp, err := a.client.Get(a.opts.JWKSURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var j jwks
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, k := range j.Keys {
		kid, _ := k["kid"].(string)
		if kid == "" {
			continue
		}
		n, _ := base64urlInt(k["n"])
		e, _ := base64urlInt(k["e"])
		if n == nil || e == nil {
			continue
		}
		a.keys[kid] = &rsa.PublicKey{N: n, E: int(e.Int64())}
	}
	return nil
}

func base64urlInt(v any) (*big.Int, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("not a string")
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(b), nil
}

func parsePublicKeyPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

// EnvOptions builds Options from environment variables.
func EnvOptions(prefix string) Options {
	return Options{
		Issuer:       os.Getenv(prefix + "_ISSUER"),
		JWKSURL:      os.Getenv(prefix + "_JWKS_URL"),
		PublicKeyPEM: os.Getenv(prefix + "_PUBLIC_KEY_PEM"),
		SkipVerify:   os.Getenv(prefix+"_SKIP_VERIFY") == "true",
	}
}
