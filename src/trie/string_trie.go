// This file contains the types and functions used to model a trie. Unlike a
// regular trie which splits on each character, this trie splits on a custom
// character, such as '/' or '.'. This allows us to model a URL path hierarchy
// and domain hierarchies. The trie only supports inserts and lookups and
// supports both static and dynamic matches. When a dynamic path component is
// included within a path, the corresponding key within the trie is marked as a
// "wildcard", and the leaf value that holds the inserted value contains a
// "dynamic matcher".
//
// Lookup performs a BFS to find all eligible candidates for the rewrite. The
// way insertion is done means that each node only has at most 1 "wildcard"
// child, since all wildcard components are grouped together. Once the input
// path has been traversed, the remaining eligible candidates are examined.
// These are all candidates that do match the key, with differing degrees of
// specificity.
//
// A match is most specific if it is a completely static match - no dynamic
// components at all. Trie construction guarantees that there is only ever at
// most one of these matches. Dynamic matches are compared using three degrees
// of specificity. A dynamic match, A, is strictly more specific than another,
// B, if A has strictly less dynamic components than B. If A and B have the
// same number of dynamic components, then we compare the required prefixes and
// suffixes of the dynamic components. Out of A and B, whichever has more
// required prefixes and suffixes is strictly more specific. If they have the
// same number of prefixes and suffixes, whichever has more _leading_ static
// components is more specific e.g. /foo/bar/* is more specific than /foo/*/baz.
// If they still cannot be separated, then whichever appeared first in lookup
// is chosen. The order of lookup appearance depends on insertion order.
//
// Once a match is chosen, we perform additional extraction of useful
// information for the caller, which is done using an extraction function
// provided at instantiation time. This is useful (and therefore only
// performed) for dynamic matches. For example, if we assume a trie that is
// used to route API requests to their appropriate handler, we would typically
// want to access the dynamic path parameters within the handler as they
// normally represent useful information such as an ID.
//
// https://www.geeksforgeeks.org/dsa/trie-insert-and-search/

package trie

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sktylr/routeit/cmp"
)

// This regex matches against a string that stats with a :, followed by at
// least 1 alphanumeric (or _ or -) character, optionally followed by 0, 1 or 2
// pipes (|) that can contain alphanumeric characters after each pipe. It is
// used to represent the expected syntax for dynamic path components. These
// components must contain a name (the characters between : and the first |),
// and optionally require a prefix (between the first and second |) or a suffix
// (after the last |). For example, the pattern :foo|bar|qux would match
// against strings starting with "bar" and ending with "qux", and apply the
// name "foo" to those that match.
var dynamicKeyRegex = regexp.MustCompile(`^:([\w-]+)(?:\|([\w.-]*))?(?:\|([\w.-]*))?$`)

// A [PathExtractor] is used to extract additional information from the matched
// path. This is commonly used for dynamic matches, as the user will typically
// want to extract the dynamic components of the matches and use it in some way.
type PathExtractor[I, O any] interface {
	// This is called when the path that is matched is entirely static. As a
	// result, there is no special extraction needed.
	NewFromStatic(val *I) *O

	// This is invoked whenever the matches path contains at the very least 1
	// dynamic component. The regex passes will match against the path and use
	// the same names, prefixes and suffixes used when inserting the node into
	// the trie.
	NewFromDynamic(val *I, parts []string, re *regexp.Regexp, indices map[string]int) *O
}

// A [StringTrie] is similar to a regular trie, except the split happens on the
// custom character specified when initialising the trie. This allows it to be
// used for a variety of string searching techniques, and it supports both
// static and dynamic matches.
type StringTrie[I any, O any] struct {
	root      *stringNode[I]
	split     rune
	extractor PathExtractor[I, O]
}

type stringNode[T any] struct {
	key      *cmp.ExactOrWildcard
	value    *stringTrieValue[T]
	children []*stringNode[T]
}

type stringTrieValue[T any] struct {
	dm  *dynamicMatcher
	val *T
}

// A dynamic matcher is used in value nodes to signify that there is at least
// one component of that node's path that is dynamic in nature. This stores the
// regex for the path, which is a named character matcher, and also the total
// number of dynamic components and the position of the first occurrence of a
// dynamic component in the path, which are both used for prioritisation.
type dynamicMatcher struct {
	re                *regexp.Regexp
	indices           map[string]int
	total             int
	first             int
	prefixSuffixCount int
}

func NewStringTrie[I, O any](split rune, extractor PathExtractor[I, O]) *StringTrie[I, O] {
	return &StringTrie[I, O]{root: &stringNode[I]{}, split: split, extractor: extractor}
}

func newStringKey(part string) *cmp.ExactOrWildcard {
	isWildcard, prefix, suffix := splitDynamicPrefixAndSuffix(part)
	if !isWildcard {
		return cmp.NewExactMatcher(part)
	}

	return cmp.NewWildcardMatcher(prefix, suffix)
}

func (t *StringTrie[I, O]) Find(path []string) (*O, bool) {
	if t.root == nil {
		return nil, false
	}

	eligible := []*stringNode[I]{t.root}
	for _, seg := range path {
		eligibleChildren := []*stringNode[I]{}
		found := false
		for _, current := range eligible {
			for _, child := range current.children {
				if child.key.Matches(seg) {
					eligibleChildren = append(eligibleChildren, child)
					found = true
				}
			}
		}
		if !found {
			return nil, false
		}
		eligible = eligibleChildren
	}

	var found *stringNode[I]
	for _, e := range eligible {
		if e.value == nil {
			// The eligible candidate is not a value node (i.e. it has children)
			continue
		}
		if e.HigherPriority(found) {
			found = e
		}
	}

	if found == nil || found.value == nil {
		return nil, false
	}

	// We omit the nil check on the inner value since by construction it should
	// always be populated.
	if found.value.dm == nil {
		return t.extractor.NewFromStatic(found.value.val), true
	}
	// TODO: this will have to be adapted for routing versus rewrites
	// TODO: this will need to work without joining - i.e. through returning a URI or similar for rewrites
	val := t.extractor.NewFromDynamic(found.value.val, path, found.value.dm.re, found.value.dm.indices)
	return val, true
}

func (t *StringTrie[I, O]) Insert(path string, value *I) {
	if t.root == nil {
		t.root = &stringNode[I]{}
	}

	current := t.root
	for seg := range strings.SplitSeq(strings.TrimPrefix(path, string(t.split)), string(t.split)) {
		current = current.GetOrCreateChild(seg)
	}

	if current.value != nil && current.value.val != value {
		panic(fmt.Errorf(`found multiple conflicting dynamic routes for %#q - found "%+v" and "%+v"`, path, current.value.val, value))
	}

	dynamicMatcher := dynamicPathToMatcher(path, t.split)
	if dynamicMatcher == nil {
		current.value = &stringTrieValue[I]{val: value}
		return
	}

	current.value = &stringTrieValue[I]{val: value, dm: dynamicMatcher}
}

func (n *stringNode[T]) GetOrCreateChild(key string) *stringNode[T] {
	wildcard, prefix, suffix := splitDynamicPrefixAndSuffix(key)
	var best *stringNode[T]
	for _, child := range n.children {
		if child.key.SameExact(key) {
			// We don't use the wildcard comparison here, otherwise we would
			// match all static paths against dynamic paths, causing some nodes
			// to be overwritten depending on the order of insertions.
			return child
		} else if wildcard && child.key.SameWildcard(prefix, suffix) {
			// Doing this ensures that we have 1 dynamic mode per (prefix,
			// suffix) combination per group of children. We don't care about
			// the name used for the dynamic match here - only the required
			// prefixes and suffixes. Since the vast majority of dynamic
			// matches will be inserted without prefixes or suffixes, this
			// generally means that each node will have at most 1 dynamic node
			// in its group of children. We do this to keep the trie sparse,
			// which helps with repeated lookups.
			best = child
		}
	}
	if best != nil {
		return best
	}
	newChild := &stringNode[T]{key: newStringKey(key)}
	n.children = append(n.children, newChild)
	return newChild
}

// Determines whether a node has strictly higher priority than another node. If
// n is a static node (i.e. no parts of its path are dynamic), then it has
// higher priority than anything else. If n is not static and other is, then
// other takes priority. If both are dynamic, then we compare their dynamic
// components. If n has strictly less dynamic components than other, n takes
// priority. If they have the same, we compare the specificity of the dynamic
// components. A dynamic component that requires a prefix or suffix is more
// specific than one that does not. So dynamic paths featuring more required
// prefixes or suffixes are strictly more specific. If the number of prefixes
// and suffixes is the same, and the number of dynamic path components is the
// same, then we compare the first appearance of a dynamic path component. A
// later first dynamic path component is strictly more specific, since it
// features more leading static components, which must match exactly by their
// nature.
func (n *stringNode[T]) HigherPriority(other *stringNode[T]) bool {
	if other == nil {
		return true
	}
	if n.value.dm == nil {
		// An exact match is always the highest priority.
		return true
	}
	if other.value.dm == nil {
		return false
	}
	return n.value.dm.HigherPriority(other.value.dm)
}

func (dm *dynamicMatcher) HigherPriority(other *dynamicMatcher) bool {
	if dm.total < other.total {
		return true
	}
	if dm.total > other.total {
		return false
	}
	if dm.prefixSuffixCount > other.prefixSuffixCount {
		return true
	} else if dm.prefixSuffixCount < other.prefixSuffixCount {
		return false
	}
	return dm.first > other.first
}

// Constructs a dynamic matcher for a given path, returning nil if the path has
// no dynamic components. This includes building a named regex that can be used
// to extract the path parameters of the request once matched.
func dynamicPathToMatcher(path string, sep rune) *dynamicMatcher {
	if !strings.Contains(path, ":") {
		return nil
	}

	// TODO: some of the leading slash stuff makes this more confusing than it should be

	frequencies, indices := map[string]int{}, map[string]int{}
	first, total, prefixSuffixCount := int(^uint(0)>>1), 0, 0
	var sb strings.Builder
	sb.WriteRune('^')
	trimmed := strings.TrimPrefix(path, string(sep))
	for i, seg := range strings.Split(trimmed, string(sep)) {
		sb.WriteRune(sep)
		if i == 0 {
			sb.WriteRune('?')
		}
		if !strings.HasPrefix(seg, ":") {
			sb.WriteString(seg)
		} else {
			// We have a segment that is ":name", optionally followed by 0, 1
			// or 2 pipes (|). Each pipe is succeeded by an alphanumeric string
			// of length 0+. In this context, we only care about the `name`
			// component, but we should validate that the shape of the entire
			// string is valid, otherwise we panic. We take the `name` and
			// convert it into a named character group that allows ASCII
			// characters, stopping at the first occurrence of the separator.
			matches := dynamicKeyRegex.FindStringSubmatch(seg)
			if matches == nil {
				panic(fmt.Errorf("invalid dynamic matcher sequence: offender=%#q, full sequence=%#q", seg, path))
			}
			name, prefix, suffix := matches[1], matches[2], matches[3]
			sb.WriteString(fmt.Sprintf(`(?P<%s>%s[^%c]+%s)`, name, prefix, sep, suffix))
			if prefix != "" {
				prefixSuffixCount++
			}
			if suffix != "" {
				prefixSuffixCount++
			}
			total++
			if i < first {
				first = i
			}
			count, exists := frequencies[name]
			if !exists {
				frequencies[name] = 1
			} else {
				frequencies[name] = count + 1
			}
			indices[name] = i
		}
	}
	sb.WriteRune('$')

	for k, v := range frequencies {
		if v != 1 {
			panic(fmt.Sprintf("duplicate entries in same route for name %#q", k))
		}
	}

	if total == 0 {
		// Although the path contained at least 1 colon, it was not in the
		// right place to signify a dynamic match, so we treat it statically
		return nil
	}

	re := regexp.MustCompile(sb.String())
	return &dynamicMatcher{re: re, total: total, first: first, prefixSuffixCount: prefixSuffixCount, indices: indices}
}

func splitDynamicPrefixAndSuffix(in string) (bool, string, string) {
	if !strings.HasPrefix(in, ":") {
		return false, "", ""
	}

	parts := strings.Split(in, "|")
	if len(parts) > 2 {
		return true, parts[1], parts[2]
	} else if len(parts) > 1 {
		return true, parts[1], ""
	}
	return true, "", ""
}
