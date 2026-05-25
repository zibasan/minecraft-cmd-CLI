package core

import "sort"

// Levenshtein calculates the Levenshtein distance between two strings.
// It is rune-aware to correctly handle multi-byte characters.
func Levenshtein(a, b string) int {
	rA := []rune(a)
	rB := []rune(b)
	la, lb := len(rA), len(rB)

	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if rA[i-1] == rB[j-1] {
				cost = 0
			}
			dp[i][j] = minInt(
				dp[i-1][j]+1,      // Deletion
				minInt(
					dp[i][j-1]+1,      // Insertion
					dp[i-1][j-1]+cost, // Substitution
				),
			)
		}
	}
	return dp[la][lb]
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type score struct {
	val  string
	dist int
}

// SuggestSimilar returns up to `max` items from the `pool` that are closest to `input`
// based on Levenshtein distance.
func SuggestSimilar(input string, pool []string, max int) []string {
	if len(pool) == 0 {
		return nil
	}

	scores := make([]score, len(pool))
	for i, p := range pool {
		scores[i] = score{
			val:  p,
			dist: Levenshtein(input, p),
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].dist < scores[j].dist
	})

	if max > len(scores) {
		max = len(scores)
	}

	results := make([]string, max)
	for i := 0; i < max; i++ {
		results[i] = scores[i].val
	}
	return results
}
