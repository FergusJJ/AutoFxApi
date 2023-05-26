package helper

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func RemoveElem(s []string, str string) []string {
	newSlice := []string{}
	for _, val := range s {
		if val != str {
			newSlice = append(newSlice, val)
		}
	}
	return newSlice
}
