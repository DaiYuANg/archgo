package internal

// JSONResultString converts marshal output into a printable string with fallback.
func JSONResultString(data []byte, err error, fallback string) string {
	if err != nil {
		return fallback
	}
	return string(data)
}
