package pagetoken

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Encode serializes token to JSON and encodes as "prefix:base64".
func Encode(prefix string, token any) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("marshal page token: %w", err)
	}
	return fmt.Sprintf("%s:%s", prefix, base64.RawURLEncoding.EncodeToString(data)), nil
}

// Decode decodes a "prefix:base64" string into token.
// Returns an error if the prefix doesn't match or the token is invalid.
func Decode(raw string, prefix string, token any) error {
	if raw == "" {
		return nil
	}
	// Extract prefix
	idx := strings.Index(raw, ":")
	if idx == -1 {
		return fmt.Errorf("invalid page token format: missing prefix")
	}
	if raw[:idx] != prefix {
		return fmt.Errorf("page token prefix mismatch: expected %q, got %q", prefix, raw[:idx])
	}
	data, err := base64.RawURLEncoding.DecodeString(raw[idx+1:])
	if err != nil {
		return fmt.Errorf("decode page token: %w", err)
	}
	if err := json.Unmarshal(data, token); err != nil {
		return fmt.Errorf("unmarshal page token: %w", err)
	}
	return nil
}
