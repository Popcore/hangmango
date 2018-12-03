package utils

// Contains returns true if the heystack slice contains needle.
func Contains(heystack []string, needle string) bool {
	for _, c := range heystack {
		if c == needle {
			return true
		}
	}
	return false
}
