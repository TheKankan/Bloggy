package validator

func IsValidUsername(s string) bool {
	const maxInput = 20

	if len(s) == 0 || len(s) > maxInput {
		return false
	}

	// No space
	for _, r := range s {
		if r == ' ' {
			return false
		}
	}

	return true
}

func IsValidPassword(s string) bool {
	const minInput = 4
	const maxInput = 100

	return len(s) >= minInput && len(s) <= maxInput
}
