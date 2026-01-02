package ai

// TruncateKeyForID shortens an API key for display purposes.
func TruncateKeyForID(key string) string {
	if len(key) > 8 {
		return key[:4] + "..." + key[len(key)-4:]
	}
	return key
}
