package validation

// Text Check the byte-length limits of a previously trimmed text value
func Text(value string, maxBytes int) bool {
	return len(value) >= 3 && len(value) <= maxBytes
}
