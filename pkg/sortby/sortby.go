// Package sortby holds comparison helpers for ordering user-facing lists.
package sortby

import (
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// Natural returns a comparison function that orders names the way a person reads
// them: runs of digits compare as numbers, so "Metal 699" sorts above "Metal 70",
// and letters follow Swedish alphabetical order, so å, ä and ö come after z. It
// returns a negative number when a sorts before b and a positive number otherwise.
//
// A collator holds scratch state, so the returned function is not safe for
// concurrent use. Make one per sort.
func Natural() func(a, b string) int {
	collator := collate.New(language.Swedish, collate.Numeric)
	return func(a, b string) int {
		if c := collator.CompareString(a, b); c != 0 {
			return c
		}
		// Names that read the same, "Metal 07" and "Metal 7", still need one fixed
		// order or the list shuffles between requests.
		return strings.Compare(a, b)
	}
}
