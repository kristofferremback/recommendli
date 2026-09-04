package sortby_test

import (
	"sort"
	"testing"

	"github.com/kristofferostlund/recommendli/pkg/sortby"
)

func TestNaturalSortsNumbersAsNumbers(t *testing.T) {
	compare := sortby.Natural()
	names := []string{"Metal 10", "Metal 600", "Metal 9", "Metal 70", "Metal", "Metal 2"}
	want := []string{"Metal", "Metal 2", "Metal 9", "Metal 10", "Metal 70", "Metal 600"}

	sort.SliceStable(names, func(i, j int) bool { return compare(names[i], names[j]) < 0 })

	assertOrder(t, want, names)
}

func TestNaturalIgnoresCase(t *testing.T) {
	compare := sortby.Natural()
	names := []string{"Zouk", "ambient", "Blues", "aphex"}
	want := []string{"ambient", "aphex", "Blues", "Zouk"}

	sort.SliceStable(names, func(i, j int) bool { return compare(names[i], names[j]) < 0 })

	assertOrder(t, want, names)
}

func TestNaturalReversedPutsTheBiggestNumberFirst(t *testing.T) {
	compare := sortby.Natural()
	names := []string{
		"Metal 70 - old one",
		"Metal 699 - almost there",
		"Metal 9 - early days",
		"Metal 670 - uselessfulness",
	}
	want := []string{
		"Metal 699 - almost there",
		"Metal 670 - uselessfulness",
		"Metal 70 - old one",
		"Metal 9 - early days",
	}

	sort.SliceStable(names, func(i, j int) bool { return compare(names[i], names[j]) > 0 })

	assertOrder(t, want, names)
}

func TestNaturalPutsSwedishLettersAfterZ(t *testing.T) {
	compare := sortby.Natural()
	names := []string{"Ölhäv", "Zouk", "Ängar", "Åka skidor", "Ambient"}
	want := []string{"Ambient", "Zouk", "Åka skidor", "Ängar", "Ölhäv"}

	sort.SliceStable(names, func(i, j int) bool { return compare(names[i], names[j]) < 0 })

	assertOrder(t, want, names)
}

func TestNaturalComparesPairs(t *testing.T) {
	compare := sortby.Natural()
	for _, tc := range []struct {
		name   string
		a, b   string
		expect string // "before", "after" or "same"
	}{
		{"numbers beat text length", "Metal 9", "Metal 10", "before"},
		{"digits sort ahead of letters", "9 Metal", "Metal", "before"},
		{"several number runs", "Set 2 Part 10", "Set 2 Part 9", "after"},
		{"a number past int64", "Mix 99999999999999999999", "Mix 99999999999999999998", "after"},
		{"zero padding reads the same", "Metal 07", "Metal 7", "before"}, // tie broken by the raw text
		{"identical names", "Metal 7", "Metal 7", "same"},
		{"prefix comes first", "Metal", "Metal Ballads", "before"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compare(tc.a, tc.b)
			if outcome(got) != tc.expect {
				t.Errorf("Natural(%q, %q) = %d, want %s", tc.a, tc.b, got, tc.expect)
			}
			if reverse := compare(tc.b, tc.a); outcome(reverse) != opposite(tc.expect) {
				t.Errorf("Natural(%q, %q) = %d, want %s", tc.b, tc.a, reverse, opposite(tc.expect))
			}
		})
	}
}

func outcome(c int) string {
	switch {
	case c < 0:
		return "before"
	case c > 0:
		return "after"
	default:
		return "same"
	}
}

func opposite(expect string) string {
	switch expect {
	case "before":
		return "after"
	case "after":
		return "before"
	default:
		return "same"
	}
}

func assertOrder(t *testing.T, want, got []string) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("got %d names, want %d", len(got), len(want))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
