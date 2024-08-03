package helpers

func WeWant[T ~string, U ~string](want T, have []U) bool {
	wantStr := string(want)

	if want == "snaptime" || want == "crtime" || len(have) == 0 {
		return true
	}

	for _, thing := range have {
		if string(thing) == wantStr {
			return true
		}
	}

	return false
}
