package sortby

import (
	"regexp"
)

type SortFunc func(i, j int) bool

func PaddedNumbers(a, b string, numLen int, asc bool) bool {
	if asc {
		return padNumbers(a, numLen) < padNumbers(b, numLen)
	} else {
		return padNumbers(a, numLen) > padNumbers(b, numLen)
	}
}

func padNumbers(v string, numLen int) string {
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllStringFunc(v, func(s string) string {
		out := s
		for len(out) < numLen {
			out = "0" + out
		}
		return out
	})
}
