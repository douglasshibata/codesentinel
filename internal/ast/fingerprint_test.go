package ast_test

import (
	"testing"

	"github.com/codesentinel/codesentinel/internal/ast"
)

func TestStructuralFingerprintStability(t *testing.T) {
	fp1 := ast.GenerateStructuralFingerprint("SEC-001", "internal/auth/jwt.go", "ValidateToken", "token := req.Header.Get(\"Authorization\")")
	// Line whitespace drift
	fp2 := ast.GenerateStructuralFingerprint("SEC-001", "internal/auth/jwt.go", "ValidateToken", "token   :=   req.Header.Get(\"Authorization\")")

	if fp1 != fp2 {
		t.Fatalf("Expected structural fingerprints to be invariant to whitespace drift: %s != %s", fp1, fp2)
	}

	// Different ruleID should yield different hash
	fp3 := ast.GenerateStructuralFingerprint("SEC-002", "internal/auth/jwt.go", "ValidateToken", "token := req.Header.Get(\"Authorization\")")
	if fp1 == fp3 {
		t.Fatalf("Expected different fingerprint for different rule ID")
	}
}

func TestExtractEnclosingScope(t *testing.T) {
	code := `package service

func ProcessPayment() {
	doSomething()
}

func ValidateUser() {
	targetLine()
}
`
	scope := ast.ExtractEnclosingScope(code, 8)
	if scope != "ValidateUser" {
		t.Fatalf("Expected enclosing scope 'ValidateUser', got '%s'", scope)
	}
}
