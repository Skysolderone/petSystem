package socialauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestOIDCVerifierVerifyIDToken(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	publicKey := privateKey.Public().(*rsa.PublicKey)
	verifier := NewOIDCVerifier(OIDCVerifierConfig{
		JWKSURL:          "https://jwks.example.test/google",
		AllowedIssuers:   []string{"https://accounts.google.com", "accounts.google.com"},
		AllowedAudiences: []string{"petverse-google-client"},
		HTTPTimeout:      2 * time.Second,
		CacheTTL:         time.Minute,
		HTTPClient:       newMockJWKSClient(t, "kid-1", publicKey),
	})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":     "https://accounts.google.com",
		"aud":     "petverse-google-client",
		"sub":     "google-user-1",
		"email":   "google@example.com",
		"name":    "Google Tester",
		"picture": "https://example.com/avatar.png",
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
		"iat":     time.Now().Add(-time.Minute).Unix(),
	})
	token.Header["kid"] = "kid-1"

	rawToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	identity, err := verifier.VerifyIDToken(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("VerifyIDToken() error = %v", err)
	}
	if identity.Subject != "google-user-1" {
		t.Fatalf("identity.Subject = %q want %q", identity.Subject, "google-user-1")
	}
	if identity.Email != "google@example.com" {
		t.Fatalf("identity.Email = %q want %q", identity.Email, "google@example.com")
	}
	if identity.Name != "Google Tester" {
		t.Fatalf("identity.Name = %q want %q", identity.Name, "Google Tester")
	}
}

func TestOIDCVerifierRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	publicKey := privateKey.Public().(*rsa.PublicKey)
	verifier := NewOIDCVerifier(OIDCVerifierConfig{
		JWKSURL:          "https://jwks.example.test/apple",
		AllowedIssuers:   []string{"https://appleid.apple.com"},
		AllowedAudiences: []string{"com.petverse.ios"},
		HTTPClient:       newMockJWKSClient(t, "kid-2", publicKey),
	})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   "https://appleid.apple.com",
		"aud":   "other-app",
		"sub":   "apple-user-1",
		"email": "apple@example.com",
		"exp":   time.Now().Add(5 * time.Minute).Unix(),
	})
	token.Header["kid"] = "kid-2"

	rawToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := verifier.VerifyIDToken(context.Background(), rawToken); err == nil {
		t.Fatal("VerifyIDToken() expected audience error")
	}
}

func newMockJWKSClient(t *testing.T, keyID string, publicKey *rsa.PublicKey) *http.Client {
	t.Helper()

	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())
	body := `{"keys":[{"kty":"RSA","use":"sig","alg":"RS256","kid":"` + keyID + `","n":"` + n + `","e":"` + e + `"}]}`

	return &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type":  []string{"application/json"},
					"Cache-Control": []string{"public, max-age=60"},
				},
				Body: io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
