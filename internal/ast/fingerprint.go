package ast

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateStructuralFingerprint calcula um hash estável do achado
// considerando a regra, o caminho relativo do arquivo, o nó sintático e o escopo da função.
func GenerateStructuralFingerprint(ruleID, relativePath, enclosingScope, normalizedSnippet string) string {
	hasher := sha256.New()

	// Remove espaços em branco redundantes e comentários para imunidade cosmética
	cleanSnippet := strings.Join(strings.Fields(normalizedSnippet), " ")

	compositeKey := fmt.Sprintf("%s|%s|%s|%s",
		ruleID,
		strings.TrimSpace(relativePath),
		strings.TrimSpace(enclosingScope),
		cleanSnippet,
	)

	hasher.Write([]byte(compositeKey))
	return hex.EncodeToString(hasher.Sum(nil))
}
