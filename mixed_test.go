package stringsort

import (
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompareMixed(t *testing.T) {
	tests := []struct {
		lhs, rhs MixedKey
		want     int
	}{
		{nil, nil, 0},
		{MixedKey{}, nil, 0},
		{nil, MixedKey{}, 0},
		{MixedKey{}, MixedKey{}, 0},

		{MixedKey{{"x", 1}}, nil, 1},
		{nil, MixedKey{{"x", 1}}, -1},
		{MixedKey{{"x", 1}}, MixedKey{{"x", 1}}, 0},

		{MixedKey{{"x", 3}}, MixedKey{{"x", 2}}, 1},
		{MixedKey{{"x", 2}}, MixedKey{{"x", 2}}, 0},
		{MixedKey{{"x", 2}}, MixedKey{{"x", 3}}, -1},

		{MixedKey{{"a", 1}}, MixedKey{{"b", 1}}, -1},
		{MixedKey{{"a", 1}}, MixedKey{{"a", 1}}, 0},
		{MixedKey{{"b", 1}}, MixedKey{{"a", 1}}, 1},
		{MixedKey{{"c", 10}}, MixedKey{{"a", 1}}, 1},
	}
	for _, test := range tests {
		got := compareMixed(test.lhs, test.rhs)
		if got != test.want {
			t.Errorf("compareMixed(%v, %v): got %v, want %v", test.lhs, test.rhs, got, test.want)
		}
	}
}

func TestParseMixed(t *testing.T) {
	tests := []struct {
		input string
		want  MixedKey
	}{
		{"", nil},
		{"foo", MixedKey{{"foo", 0}}},
		{"foo 42", MixedKey{{"foo ", 42}}},
		{"101", MixedKey{{"", 101}}},
		{"alpha25bravo-3", MixedKey{{"alpha", 25}, {"bravo-", 3}}},
		{"101 dalmatians", MixedKey{{"", 101}, {" dalmatians", 0}}},
	}
	opt := cmp.AllowUnexported(nspan{})
	for _, test := range tests {
		got := ParseMixed(test.input)
		if diff := cmp.Diff(test.want, got, opt); diff != "" {
			t.Errorf("ParseMixed(%q): (-want, +got):\n%s", test.input, diff)
		}
	}
}

func TestByMixedKey(t *testing.T) {
	// The input slice must have the expected order.
	input := []string{
		// needles with leading digits
		"9foxtrot",
		"31 whisky tango foxtrot 9",
		"31 whisky tango foxtrot 89",
		"81foxtrot",
		"219 whsky tango foxtrot 9",
		"762foxtrot",
		"762foxtrot 9",
		"762foxtrot 10",

		// needles without leading digits
		"alpha 1 bravo 32",
		"alpha 10 bravo 19",
		"bravo 3 charlie",
		"bravo 4 xray",
		"charlie",
		"charlie52",
		"charlie300",

		// needles that compare equal but are not identical
		"echo001",
		"echo01",
		"echo1",
	}

	// As a sanity check on the test, verify that the input is not the same as a
	// lexicographic sort on the same strings. This ensures the rest of the test
	// is actually exercising the code.
	cp := copyStrings(input)
	sort.Strings(cp)
	if cmp.Equal(cp, input) {
		t.Fatalf("Test failed: input is already in lexicographic order:\n%s",
			strings.Join(input, "\n"))
	}

	// To exercise the code, randomly permute the input and verify that sorting
	// it brings us back the desired order.
	for i := 0; i < 50; i++ {
		cp := copyStrings(input)
		rand.Shuffle(len(cp), func(i, j int) {
			cp[i], cp[j] = cp[j], cp[i]
		})

		got := copyStrings(cp)
		sort.Sort(ByMixedKey(got))
		if diff := cmp.Diff(input, got); diff != "" {
			t.Errorf("ByMixedKey(%+q): (-want, +got):\n%s", cp, diff)
		}
	}
}

func copyStrings(ss []string) []string {
	cp := make([]string, len(ss))
	copy(cp, ss)
	return cp
}
