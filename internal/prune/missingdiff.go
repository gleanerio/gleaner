package prune

// findMissingElements from chatGPT
// You'll have my job one done wont you!
func findMissing(a, b []string) []string {
	// Create a map to store the elements of ga.
	gaMap := make(map[string]bool)
	for _, s := range b {
		gaMap[s] = true
	}

	// Iterate through a and add any elements that are not in b to the result slice.
	var result []string
	for _, s := range a {
		if !gaMap[s] {
			result = append(result, s)
		}
	}

	return result
}
