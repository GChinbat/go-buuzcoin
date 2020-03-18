package trie

func longestCommonPrefixLength(s1, s2 string) int {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	result := 0
	for result < minLen && s1[result] == s2[result] {
		result++
	}

	return result
}
