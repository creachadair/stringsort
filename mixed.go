// Package stringsort provides support code for sorting strings.
//
// Mixed Keys
//
// Ordinarily strings are sorted lexicographically by character.  This is
// simple and consistent, but when applied to UI elements can be unintuitive
// for users. For example, lexicgraphically sorting a list of filenames will
// produce an order like
//
//   file-1.png
//   file-10.png
//   file-2.png
//
// That is, "file 2" is listed after "file 10". One way to address this is to
// treat runs of digits differently in comparison: Instead of comparing them
// digit-by-digit, treat the entire run as a single value.
//
// The MixedKey type supports this representation, representing a string as a
// sequence of "spans", each consisting of a non-digit string followed by an
// integer value corresponding to a digit string. Lexicographic comparison of
// these keys will preserve the intuitive ordering of digit sequences.
//
// This approach emulates the ordering used by the macOS Finder for file names.
//
package stringsort

import "sort"

// ByMixedKey returns a sorter that orders ss non-decreasing by mixed key. The
// keys are precomputed at the point of construction.
//
// Note that non-identical strings may have equal mixed keys, consider for
// example "xyzzy1" and "xyzzy01". To ensure a deterministic order, ties on key
// order are broken using the lexicgraphic order of the original strings.
func ByMixedKey(ss []string) sort.Interface {
	kp := byMixedKey{
		ss:   ss,
		keys: make([]MixedKey, len(ss)),
	}
	for i, s := range ss {
		kp.keys[i] = ParseMixed(s)
	}
	return kp
}

// byMixedKey implements sort.Interface using mixed keys.
type byMixedKey struct {
	ss   []string   // the original slice to be sorted
	keys []MixedKey // keys corresponding to ss
}

func (b byMixedKey) Len() int { return len(b.ss) }

func (b byMixedKey) Less(i, j int) bool {
	v := compareMixed(b.keys[i], b.keys[j])
	if v == 0 {
		// Break ties using lexicographic order, to ensure deterministic output.
		return b.ss[i] < b.ss[j]
	}
	return v < 0
}

func (b byMixedKey) Swap(i, j int) {
	b.ss[i], b.ss[j] = b.ss[j], b.ss[i]         // permute the strings
	b.keys[i], b.keys[j] = b.keys[j], b.keys[i] // update their keys
}

// A MixedKey is a lexicographic sort key for a string that partitions it into
// paired runs of non-digits and decimal digits. The runs of digits are
// interpreted as integer values for comparison.
//
// For example, the string "alpha25bravo-3" generates the mixed key:
//
//      ("alpha", 25) ("bravo-", 3)
//
// while the string "101 dalmatians" generates the mixed key:
//
//      ("", 101) (" dalmatians", 0)
//
type MixedKey []nspan

// ParseMixed parses s into a MixedKey.
func ParseMixed(s string) MixedKey {
	var out MixedKey

	i, end := 0, 0
	for i < len(s) {
		// Scan for a digit
		ch := s[i]
		if ch < '0' || ch > '9' {
			i++
			continue
		}

		// Having found a digit, start a new span with the run prior to the
		// digit.  Consume digits until a non-digit or end-of-string.  Note the
		// prior span may be empty, if the string begins with digits.
		cur := nspan{run: s[end:i], n: int(ch - '0')}
		i++
		for i < len(s) {
			ch := s[i]
			if ch < '0' || ch > '9' {
				break
			}
			cur.n = 10*cur.n + int(ch-'0')
			i++
		}
		out = append(out, cur)
		end = i
	}

	// Ensure a non-empty trailing run is captured.
	if end < i {
		out = append(out, nspan{run: s[end:i]})
	}
	return out
}

func compareInt(a, b int) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}

type nspan struct {
	run string
	n   int
}

func compareNspan(a, b nspan) int {
	if a.run == b.run {
		return compareInt(a.n, b.n)
	} else if a.run < b.run {
		return -1
	}
	return 1
}

func compareMixed(a, b MixedKey) int {
	i := 0
	for i < len(a) && i < len(b) {
		if c := compareNspan(a[i], b[i]); c != 0 {
			return c
		}
		i++
	}
	return compareInt(len(a), len(b))
}
