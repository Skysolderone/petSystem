package socialauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	GoogleJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
	AppleJWKSURL  = "https://appleid.apple.com/auth/keys"
)

type Identity struct {
	Subject   string
	Email     string
	Name      string
	Picture   string
	Issuer    string
	Audience  []string
	ExpiresAt time.Time
}

type Verifier interface {
	VerifyIDToken(ctx context.Context, token string) (*Identity, error)
}

type OIDCVerifierConfig struct {
	JWKSURL          string
	AllowedIssuers   []string
	AllowedAudiences []string
	HTTPTimeout      time.Duration
	CacheTTL         time.Duration
	HTTPClient       *http.Client
}

type OIDCVerifier struct {
	allowedIssuers   []string
	allowedAudiences []string
	httpClient       *http.Client
	jwksURL          string
	cacheTTL         time.Duration

	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	cacheExpiry time.Time
}

type identityClaims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	KeyType  string `json:"kty"`
	KeyID    string `json:"kid"`
	Use      string `json:"use"`
	Alg      string `json:"alg"`
	Modulus  string `json:"n"`
	Exponent string `json:"e"`
}

func NewOIDCVerifier(cfg OIDCVerifierConfig) *OIDCVerifier {
	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	cacheTTL := cfg.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &OIDCVerifier{
		allowedIssuers:   normalizeValues(cfg.AllowedIssuers),
		allowedAudiences: normalizeValues(cfg.AllowedAudiences),
		httpClient:       httpClient,
		jwksURL:          strings.TrimSpace(cfg.JWKSURL),
		cacheTTL:         cacheTTL,
		keys:             map[string]*rsa.PublicKey{},
	}
}

func (v *OIDCVerifier) VerifyIDToken(ctx context.Context, rawToken string) (*Identity, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, fmt.Errorf("identity token is required")
	}
	if len(v.allowedAudiences) == 0 {
		return nil, fmt.Errorf("provider audiences are not configured")
	}

	claims := &identityClaims{}
	parsed, err := jwt.NewParser().ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}

		keyID, _ := token.Header["kid"].(string)
		if keyID == "" {
			return nil, fmt.Errorf("missing key id")
		}
		return v.lookupKey(ctx, keyID)
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("invalid identity token")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("identity token subject is missing")
	}
	if !containsString(v.allowedIssuers, claims.Issuer) {
		return nil, fmt.Errorf("invalid issuer")
	}
	if !hasAnyAudience(claims.Audience, v.allowedAudiences) {
		return nil, fmt.Errorf("invalid audience")
	}
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("identity token has expired")
	}

	return &Identity{
		Subject:   claims.Subject,
		Email:     claims.Email,
		Name:      claims.Name,
		Picture:   claims.Picture,
		Issuer:    claims.Issuer,
		Audience:  audienceValues(claims.Audience),
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (v *OIDCVerifier) lookupKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	if key, ok := v.cachedKey(keyID); ok {
		return key, nil
	}
	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}
	if key, ok := v.cachedKey(keyID); ok {
		return key, nil
	}
	return nil, fmt.Errorf("signing key not found")
}

func (v *OIDCVerifier) cachedKey(keyID string) (*rsa.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if time.Now().After(v.cacheExpiry) {
		return nil, false
	}
	key, ok := v.keys[keyID]
	return key, ok
}

func (v *OIDCVerifier) refreshKeys(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}

	response, err := v.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("jwks request failed with status %d", response.StatusCode)
	}

	var document jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, key := range document.Keys {
		publicKey, err := decodeRSAKey(key)
		if err != nil {
			continue
		}
		keys[key.KeyID] = publicKey
	}
	if len(keys) == 0 {
		return fmt.Errorf("jwks response contained no valid rsa keys")
	}

	ttl := parseCacheControlTTL(response.Header.Get("Cache-Control"))
	if ttl <= 0 {
		ttl = v.cacheTTL
	}

	v.mu.Lock()
	v.keys = keys
	v.cacheExpiry = time.Now().Add(ttl)
	v.mu.Unlock()

	return nil
}

func decodeRSAKey(key jwkKey) (*rsa.PublicKey, error) {
	if key.KeyType != "RSA" || key.KeyID == "" || key.Modulus == "" || key.Exponent == "" {
		return nil, fmt.Errorf("invalid jwk")
	}

	modulusBytes, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	if err != nil {
		return nil, err
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	if err != nil {
		return nil, err
	}

	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)
	if modulus.Sign() <= 0 || exponent.Sign() <= 0 {
		return nil, fmt.Errorf("invalid rsa key")
	}

	return &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}, nil
}

func parseCacheControlTTL(value string) time.Duration {
	for _, directive := range strings.Split(value, ",") {
		part := strings.TrimSpace(directive)
		if !strings.HasPrefix(part, "max-age=") {
			continue
		}

		rawAge := strings.TrimPrefix(part, "max-age=")
		seconds, err := time.ParseDuration(rawAge + "s")
		if err == nil && seconds > 0 {
			return seconds
		}
	}
	return 0
}

func containsString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasAnyAudience(tokenAudiences jwt.ClaimStrings, allowed []string) bool {
	for _, audience := range tokenAudiences {
		if containsString(allowed, audience) {
			return true
		}
	}
	return false
}

func audienceValues(values jwt.ClaimStrings) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

func normalizeValues(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}
