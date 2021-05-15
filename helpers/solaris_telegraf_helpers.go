package helpers

func WeWant(want string, have []string) bool {
	if len(have) == 0 {
		return true
	}

	for _, thing := range have {
		if thing == want {
			return true
		}
	}

	return false
}
