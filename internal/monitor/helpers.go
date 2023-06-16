package monitor

import (
	"log"
	"regexp"
	"strings"
)

func parseField(str string) string {
	re := regexp.MustCompile(`(.+)(\[.+\])`) // Regular expression pattern

	match := re.FindStringSubmatch(str) // Find the first match in the string

	if len(match) == 3 {
		return match[1] // Return the text before the square brackets
	}

	return "" // Return an empty string if no match is found

}

func splitTypeVal(typeValStr string) (string, string) {
	typeValMap := strings.Split(typeValStr, ":")
	if len(typeValMap) != 2 {
		log.Fatalf(`%s is formatted incorrectly`, typeValStr)
	}
	return typeValMap[0], typeValMap[1]
}
