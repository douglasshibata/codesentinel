package privacy_test

import (
	"strings"
	"testing"

	"github.com/codesentinel/codesentinel/internal/privacy"
)

func TestRedactSensitiveData(t *testing.T) {
	diff := `
+ apiKey := "secret_key_1234567890abcdef"
+ token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgN"
+ privKey := "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0\n-----END RSA PRIVATE KEY-----"
`
	sanitized := privacy.RedactSensitiveData(diff)

	if strings.Contains(sanitized, "secret_key_1234567890abcdef") {
		t.Fatalf("API key was not redacted")
	}
	if strings.Contains(sanitized, "eyJhbGciOiJIUzI1Ni") {
		t.Fatalf("JWT was not redacted")
	}
	if strings.Contains(sanitized, "BEGIN RSA PRIVATE KEY") {
		t.Fatalf("RSA Private key was not redacted")
	}

	if !strings.Contains(sanitized, "[REDACTED_CREDENTIAL]") {
		t.Fatalf("Missing [REDACTED_CREDENTIAL] placeholder")
	}
	if !strings.Contains(sanitized, "[REDACTED_JWT_TOKEN]") {
		t.Fatalf("Missing [REDACTED_JWT_TOKEN] placeholder")
	}
	if !strings.Contains(sanitized, "[REDACTED_RSA_PRIVATE_KEY]") {
		t.Fatalf("Missing [REDACTED_RSA_PRIVATE_KEY] placeholder")
	}
}
