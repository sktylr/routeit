package cmp

// The [ExactOrWildcard] struct is a union between a wildcard and an exact
// string match. A component can either be a wildcard, meaning it will match
// most inputs with given prefixes, suffixes and minimum lengths (or
// potentially all inputs), or can be an exact match, meaning it will only
// match the exact input.
type ExactOrWildcard struct {
	exact string
	wc    *wildcard
}

// This will only ever match the exact string passed here - any other input
// will not match
func NewExactMatcher(exact string) *ExactOrWildcard {
	return &ExactOrWildcard{exact: exact}
}

// This will match against a wildcard expression. The expression must start
// with the specific prefix and end with the given suffix. If both are empty,
// then it will match against any non-empty input. If either is populated, then
// it will match against "<prefix>*<suffix>", where * is 1 or more characters.
// This prevents overlapping - e.g. given the prefix and suffix "foob" and
// "bar" respectively, this would not match against "foobar" nor "foobbar", but
// would match against "foobabar".
func NewWildcardMatcher(pre, suf string) *ExactOrWildcard {
	return &ExactOrWildcard{wc: newWildcard(pre, suf)}
}

// Examines the union and determines whether the input matches
func (eow *ExactOrWildcard) Matches(cmp string) bool {
	if eow.isWildcard() {
		return eow.wc.Matches(cmp)
	}
	return eow.SameExact(cmp)
}

// Utility for determining whether this matches the exact portion of the union.
// This should typically only be used when comparing instances for creation -
// e.g. to determine whether a duplicate would be made.
func (eow *ExactOrWildcard) SameExact(cmp string) bool {
	return eow.exact == cmp
}

// Utility method to determine whether the given prefix and suffix match
// exactly the prefix and suffix of the wildcard component. This should only be
// used as part of a creation logic to evaluate whether two wildcard matchers
// are exactly equals
func (eow *ExactOrWildcard) SameWildcard(pre, suf string) bool {
	return eow.isWildcard() && eow.wc.prefix == pre && eow.wc.suffix == suf
}

func (eow *ExactOrWildcard) isWildcard() bool {
	return eow.wc != nil
}
