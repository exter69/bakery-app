package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// scriptPattern matches <script>...</script> blocks including their content.
var scriptPattern = regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)

// htmlTagPattern matches any HTML tag.
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// SanitizeString strips script blocks and HTML tags from the input string.
// It first removes <script>...</script> blocks (including content), then removes
// any remaining HTML tags, and finally trims extra whitespace.
func SanitizeString(input string) string {
	// Step 1: Remove <script>...</script> blocks and their content
	result := scriptPattern.ReplaceAllString(input, "")

	// Step 2: Remove all remaining HTML tags
	result = htmlTagPattern.ReplaceAllString(result, "")

	// Step 3: Collapse multiple spaces left by removed tags
	result = strings.Join(strings.Fields(result), " ")

	// Step 4: Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	return result
}

// sanitizeValue recursively walks a JSON-decoded value and sanitizes all strings.
func sanitizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return SanitizeString(val)
	case map[string]interface{}:
		for key, value := range val {
			val[key] = sanitizeValue(value)
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = sanitizeValue(item)
		}
		return val
	default:
		return v
	}
}

// InputSanitizer returns middleware that sanitizes all string values in JSON request bodies.
// It reads the body, parses the JSON, sanitizes string fields, re-marshals, and replaces
// the request body. Non-JSON requests pass through unchanged.
func InputSanitizer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only process requests that might have a JSON body
		if r.Body == nil || r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		// Read the body
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// If body is empty, just pass through
		if len(body) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// Parse JSON generically
		var parsed interface{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			// If it's not valid JSON, pass through as-is
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// Sanitize all string values
		sanitized := sanitizeValue(parsed)

		// Re-marshal
		newBody, err := json.Marshal(sanitized)
		if err != nil {
			// If re-marshaling fails, pass the original body
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// Replace request body
		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

		next.ServeHTTP(w, r)
	})
}
