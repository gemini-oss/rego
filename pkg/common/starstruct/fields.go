// pkg/common/starstruct/fields.go
package starstruct

import (
	"sort"
	"strings"
)

// MergeFields merges parent and subfields according to these rules:
//  1. If a baseline field has candidate subfields (numeric or non‑numeric),
//     the baseline field is omitted and the candidate subfields are inserted in its place.
//  2. (When numeric candidate subfields exist, they are sorted by their numeric suffix.)
//  3. If a candidate field does not map to a baseline field but shares a top‑level prefix,
//     insert it after the last baseline field that has that prefix.
//  4. Bare candidate fields (with no dot) are appended at the end if not in baseline.
//  5. No duplicates. Baseline ordering is preserved as the “master” order.
func MergeFields(baseline, candidate []string) []string {
	type numField struct {
		field string
		num   int
	}
	numericChildren := make(map[string][]numField)  // Suffix is numeric
	nonNumericChildren := make(map[string][]string) // Suffix is non-numeric
	var bareCandidate []string                      // No dot in the suffix

	for _, c := range candidate {
		// Find the last dot manually.
		dotIdx := -1
		for i := len(c) - 1; i >= 0; i-- {
			if c[i] == '.' {
				dotIdx = i
				break
			}
		}
		if dotIdx == -1 {
			bareCandidate = append(bareCandidate, c)
			continue
		}
		parent := c[:dotIdx]
		suffix := c[dotIdx+1:]

		// Try to convert the suffix to an integer without calling strconv.Atoi.
		numVal := 0
		isNumeric := len(suffix) > 0
		for j := 0; isNumeric && j < len(suffix); j++ {
			d := suffix[j]
			if d < '0' || d > '9' {
				isNumeric = false
				break
			}
			numVal = numVal*10 + int(d-'0')
		}
		if isNumeric {
			numericChildren[parent] = append(numericChildren[parent], numField{field: c, num: numVal})
		} else {
			nonNumericChildren[parent] = append(nonNumericChildren[parent], c)
		}
	}

	// Sort the numeric fields by their numeric suffix.
	for parent, arr := range numericChildren {
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].num < arr[j].num
		})
		numericChildren[parent] = arr
	}
	// Sort non-numeric candidate groups lexically.
	for parent, arr := range nonNumericChildren {
		sort.Strings(arr)
		nonNumericChildren[parent] = arr
	}

	//----------------------------------------------------------------------
	// 2. Build a set of baseline fields
	// ---------------------------------------------------------------------
	baselineSet := make(map[string]bool, len(baseline))
	for _, b := range baseline {
		baselineSet[b] = true
	}

	//----------------------------------------------------------------------
	// 3) Build the merged list from baseline fields.
	// ---------------------------------------------------------------------
	merged := make([]string, 0, len(baseline)+len(candidate))
	mergedSet := make(map[string]bool, len(baseline)+len(candidate))
	for _, b := range baseline {
		hasChildren := false

		// If there are numeric children for `b`, output them.
		if arr, exists := numericChildren[b]; exists && len(arr) > 0 {
			hasChildren = true
			for _, nf := range arr {
				if !mergedSet[nf.field] {
					merged = append(merged, nf.field)
					mergedSet[nf.field] = true
				}
			}
			// Remove these so they aren’t reprocessed later.
			delete(numericChildren, b)
		}

		// If there are non‑numeric children for `b`, output them as well.
		if arr, exists := nonNumericChildren[b]; exists && len(arr) > 0 {
			hasChildren = true
			for _, child := range arr {
				if !mergedSet[child] {
					merged = append(merged, child)
					mergedSet[child] = true
				}
			}
			delete(nonNumericChildren, b)
		}

		// Only output the baseline field if there are no subfields.
		if !hasChildren {
			if !mergedSet[b] {
				merged = append(merged, b)
				mergedSet[b] = true
			}
		}
	}

	//----------------------------------------------------------------------
	// 4) Insert extra fields for which the parent is not in baseline.
	// ---------------------------------------------------------------------
	extraFields := make(map[string][]string)
	// Process remaining numeric field groups.
	for parent, arr := range numericChildren {
		// Skip if parent is in baseline (shouldn’t happen now).
		if baselineSet[parent] {
			continue
		}
		prefix := parent
		if j := strings.IndexByte(parent, '.'); j != -1 {
			prefix = parent[:j]
		}
		for _, nf := range arr {
			extraFields[prefix] = append(extraFields[prefix], nf.field)
		}
	}
	// Process remaining non‑numeric candidate groups.
	for parent, arr := range nonNumericChildren {
		if baselineSet[parent] {
			continue
		}
		prefix := parent
		if j := strings.IndexByte(parent, '.'); j != -1 {
			prefix = parent[:j]
		}
		extraFields[prefix] = append(extraFields[prefix], arr...)
	}

	// Insert each extra field group immediately after the last occurrence in merged of a field whose top-level prefix matches.
	prefixes := make([]string, 0, len(extraFields))
	for p := range extraFields {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)
	for _, prefix := range prefixes {
		extraFields := extraFields[prefix]
		anchor := -1
		for i, f := range merged {
			p := f
			if j := strings.IndexByte(f, '.'); j != -1 {
				p = f[:j]
			}
			if p == prefix {
				anchor = i
			}
		}
		if anchor == -1 {
			// No anchor found: append at the end.
			for _, field := range extraFields {
				if !mergedSet[field] {
					merged = append(merged, field)
					mergedSet[field] = true
				}
			}
		} else {
			// Insert immediately after the anchor.
			filtered := make([]string, 0, len(extraFields))
			for _, field := range extraFields {
				if !mergedSet[field] {
					filtered = append(filtered, field)
					mergedSet[field] = true
				}
			}
			if len(filtered) > 0 {
				tail := append([]string(nil), merged[anchor+1:]...)
				merged = append(merged[:anchor+1], filtered...)
				merged = append(merged, tail...)
			}
		}
	}

	//----------------------------------------------------------------------
	// 5) Append any bare fields not already included.
	// ---------------------------------------------------------------------
	for _, bc := range bareCandidate {
		if !baselineSet[bc] && !mergedSet[bc] {
			merged = append(merged, bc)
			mergedSet[bc] = true
		}
	}

	return merged
}
