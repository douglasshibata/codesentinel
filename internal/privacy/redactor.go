package privacy

import (
	"regexp"
)

var (
	jwtRegex        = regexp.MustCompile(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`)
	apiKeyRegex     = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|bearer)[\s]*[=:]+[\s]*['"]?([a-zA-Z0-9_\-\.]{12,})['"]?`)
	privateKeyRegex = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]+PRIVATE KEY-----.*?-----END [A-Z ]+PRIVATE KEY-----`)
)

func RedactSensitiveData(sourceDiff string) string {
	redacted := privateKeyRegex.ReplaceAllString(sourceDiff, "[REDACTED_RSA_PRIVATE_KEY]")
	redacted = jwtRegex.ReplaceAllString(redacted, "[REDACTED_JWT_TOKEN]")
	redacted = apiKeyRegex.ReplaceAllString(redacted, "$1: [REDACTED_CREDENTIAL]")
	return redacted
}
