package cmp

// The [wildcard] struct allows for basic wildcard matching against an incoming
// key. The wildcard matches any incoming string that has the same prefix and
// suffix defined in the struct, and strictly more characters than minLen.
// Where the prefix and suffix are both empty, the wildcard matches against
// anything. If the fn is provided, then the stripped segment (no prefix or
// suffix) is passed to this function to determine whether the input matches.
// This function is only called whenever both the prefix and suffix match, and
// the segment between the prefix and suffix is of non-zero length.
type wildcard struct {
	prefix string
	suffix string
	pLen   int
	sLen   int
	fn     func(seg string) bool
}

func newWildcard(pre, suf string, fn func(seg string) bool) *wildcard {
	return &wildcard{prefix: pre, suffix: suf, pLen: len(pre), sLen: len(suf), fn: fn}
}

func (wc *wildcard) Matches(cmp string) bool {
	if wc.prefix == "" && wc.suffix == "" {
		return cmp != "" && (wc.fn == nil || wc.fn(cmp))
	}
	len := len(cmp)
	if len <= (wc.pLen + wc.sLen) {
		return false
	}
	i := len - wc.sLen
	pre, seg, suf := cmp[:wc.pLen], cmp[wc.pLen:i], cmp[i:]
	return pre == wc.prefix && suf == wc.suffix && (wc.fn == nil || wc.fn(seg))
}

func (wc *wildcard) PrefixMatches(cmp string) bool {
	return wc.prefix == cmp
}

func (wc *wildcard) SuffixMatches(cmp string) bool {
	return wc.suffix == cmp
}
