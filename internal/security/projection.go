package security

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	projectionEmailPattern = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	projectionLongDigits   = regexp.MustCompile(`\b\d{6,}\b`)
	projectionLongSecret   = regexp.MustCompile(`\b(?:[A-Fa-f0-9]{32,}|[A-Za-z0-9_\-]{32,})\b`)
)

// RedactedProjection returns a normalized, redacted projection suitable for
// plaintext search/display columns.
func RedactedProjection(text string, limit int) string {
	projection := strings.TrimSpace(text)
	if projection == "" {
		return ""
	}
	projection = projectionEmailPattern.ReplaceAllString(projection, "[email]")
	projection = projectionLongDigits.ReplaceAllString(projection, "[number]")
	projection = projectionLongSecret.ReplaceAllString(projection, "[secret]")
	projection = strings.Join(strings.Fields(projection), " ")
	if limit > 0 && len(projection) > limit {
		projection = projection[:limit]
	}
	return projection
}

// ProtectText returns the plaintext projection and, when a protector is
// configured, the encrypted original payload.
func ProtectText(protector PayloadProtector, text string, projectionLimit int) (projection string, ciphertext, nonce []byte, keyVersion int, err error) {
	if protector == nil {
		return strings.TrimSpace(text), nil, nil, 0, nil
	}
	projection = RedactedProjection(text, projectionLimit)
	if strings.TrimSpace(text) == "" {
		return projection, nil, nil, 0, nil
	}
	ciphertext, nonce, keyVersion, err = protector.EncryptPayload([]byte(text))
	if err != nil {
		return "", nil, nil, 0, err
	}
	return projection, ciphertext, nonce, keyVersion, nil
}

// ProtectJSONBundle encrypts a JSON-encoded structured payload.
func ProtectJSONBundle(protector PayloadProtector, v any) (ciphertext, nonce []byte, keyVersion int, err error) {
	if protector == nil {
		return nil, nil, 0, nil
	}
	plaintext, err := json.Marshal(v)
	if err != nil {
		return nil, nil, 0, err
	}
	if len(plaintext) == 0 {
		return nil, nil, 0, nil
	}
	ciphertext, nonce, keyVersion, err = protector.EncryptPayload(plaintext)
	if err != nil {
		return nil, nil, 0, err
	}
	return ciphertext, nonce, keyVersion, nil
}

// UnprotectJSONBundle decrypts and unmarshals a JSON payload bundle.
func UnprotectJSONBundle[T any](protector PayloadProtector, ciphertext, nonce []byte, keyVersion int) (T, error) {
	var zero T
	if protector == nil {
		return zero, nil
	}
	plaintext, err := protector.DecryptPayload(ciphertext, nonce, keyVersion)
	if err != nil {
		return zero, err
	}
	var out T
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return zero, err
	}
	return out, nil
}
