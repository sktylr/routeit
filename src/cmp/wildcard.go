package cmp

import "strings"

// The [wildcard] struct allows for basic wildcard matching against an incoming
// key. The wildcard matches any incoming string that has the same prefix and
// suffix defined in the struct, and strictly more characters than minLen.
// Where the prefix and suffix are both empty, the wildcard matches against
// anything.
type wildcard struct {
	prefix string
	suffix string
	minLen int
}

func newWildcard(pre, suf string) *wildcard {
	return &wildcard{prefix: pre, suffix: suf, minLen: len(pre) + len(suf)}
}

func (wc *wildcard) Matches(cmp string) bool {
	prefixEmpty, suffixEmpty := wc.prefix == "", wc.suffix == ""
	if prefixEmpty && suffixEmpty {
		return cmp != ""
	}
	if prefixEmpty {
		return strings.HasSuffix(cmp, wc.suffix) && len(cmp) > wc.minLen
	}
	if suffixEmpty {
		return strings.HasPrefix(cmp, wc.prefix) && len(cmp) > wc.minLen
	}
	return strings.HasPrefix(cmp, wc.prefix) && strings.HasSuffix(cmp, wc.suffix) && len(cmp) > wc.minLen
}

func (wc *wildcard) PrefixMatches(cmp string) bool {
	return wc.prefix == cmp
}

func (wc *wildcard) SuffixMatches(cmp string) bool {
	return wc.suffix == cmp
}
