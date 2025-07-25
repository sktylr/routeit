// This file contains the types and functions used to model a trie. Unlike a
// regular trie which splits on each character, this trie splits on the '/'
// character and is used to model a URL path hierarchy. The trie only supports
// inserts and lookups and supports both static and dynamic matches. When a
// dynamic path component is included within a path, the corresponding key
// within the trie is marked as a "wildcard", and the leaf value that holds the
// inserted value contains a "dynamic matcher".
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

package routeit

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

type trie[T any, E any] struct {
	root    *node[T]
	extract dynamicExtractor[T, E]
}

type node[T any] struct {
	key      *cmp.ExactOrWildcard
	value    *trieValue[T]
	children []*node[T]
}

type trieValue[T any] struct {
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
	total             int
	first             int
	prefixSuffixCount int
}

// A dynamic extractor operates on a [trieValue] to extract meaningful
// information from a matched [trie] key. For example, the two implementations
// used in this project are used to extract path parameters from the matched
// path so that they can be populated in the request, and perform dynamic URL
// rewrites using regex substitution.
type dynamicExtractor[T any, O any] func(*trieValue[T], string) O

func newTrie[T any, D any](extract dynamicExtractor[T, D]) *trie[T, D] {
	return &trie[T, D]{root: &node[T]{}, extract: extract}
}

func newKey(part string) *cmp.ExactOrWildcard {
	isWildcard, prefix, suffix := splitDynamicPrefixAndSuffix(part)
	if !isWildcard {
		return cmp.NewExactMatcher(part)
	}

	return cmp.NewWildcardMatcher(prefix, suffix)
}

func (t *trie[T, D]) Find(path string) (*T, *D, bool) {
	if t.root == nil {
		return nil, nil, false
	}

	eligible := []*node[T]{t.root}
	for i, seg := range strings.Split(path, "/") {
		if i == 0 && seg == "" {
			continue
		}
		eligibleChildren := []*node[T]{}
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
			return nil, nil, false
		}
		eligible = eligibleChildren
	}

	var found *node[T]
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
		return nil, nil, false
	}

	// We omit the nil check on the inner value since by construction it should
	// always be populated.
	d := t.extract(found.value, path)
	return found.value.val, &d, true
}

func (t *trie[T, D]) Insert(path string, value *T) {
	if t.root == nil {
		t.root = &node[T]{}
	}

	current := t.root
	for i, seg := range strings.Split(path, "/") {
		if i == 0 && seg == "" {
			continue
		}
		current = current.GetOrCreateChild(seg)
	}

	if current.value != nil && current.value.val != value {
		panic(fmt.Errorf(`found multiple conflicting dynamic routes for %#q - found "%+v" and "%+v"`, path, current.value.val, value))
	}

	dynamicMatcher := dynamicPathToMatcher(path)
	if dynamicMatcher == nil {
		current.value = &trieValue[T]{val: value}
		return
	}

	current.value = &trieValue[T]{val: value, dm: dynamicMatcher}
}

func (n *node[T]) GetOrCreateChild(key string) *node[T] {
	wildcard, prefix, suffix := splitDynamicPrefixAndSuffix(key)
	var best *node[T]
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
	newChild := &node[T]{key: newKey(key)}
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
func (n *node[T]) HigherPriority(other *node[T]) bool {
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

// Collects the path parameters of the matched path
func (v *trieValue[T]) PathParams(path string) pathParameters {
	if v.dm == nil {
		return pathParameters{}
	}

	params := pathParameters{}
	names := v.dm.re.SubexpNames()
	matches := v.dm.re.FindStringSubmatch(path)

	if matches == nil {
		// Indicates that something has gone wrong with the regex or searching.
		return params
	}

	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}
		params[name] = matches[i]
	}

	return params
}

// Performs substitution on the matched path, treating the matched value as a
// template for the substitution.
func stringSubstitution(v *trieValue[string], path string) string {
	if v.dm == nil {
		return ""
	}

	match := v.dm.re.FindStringSubmatchIndex(path)
	result := v.dm.re.ExpandString(nil, *v.val, path, match)

	return string(result)
}

// Constructs a dynamic matcher for a given path, returning nil if the path has
// no dynamic components. This includes building a named regex that can be used
// to extract the path parameters of the request once matched.
func dynamicPathToMatcher(path string) *dynamicMatcher {
	if !strings.Contains(path, ":") {
		return nil
	}

	// TODO: some of the leading slash stuff makes this more confusing than it should be

	frequencies := map[string]int{}
	first, total, prefixSuffixCount := int(^uint(0)>>1), 0, 0
	var sb strings.Builder
	sb.WriteRune('^')
	for i, seg := range strings.Split(path, "/") {
		if i == 0 && seg == "" {
			continue
		}
		sb.WriteRune('/')
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
			// characters, stopping at the first /.
			matches := dynamicKeyRegex.FindStringSubmatch(seg)
			if matches == nil {
				panic(fmt.Errorf("invalid dynamic matcher sequence: offender=%#q, full sequence=%#q", seg, path))
			}
			name, prefix, suffix := matches[1], matches[2], matches[3]
			sb.WriteString(fmt.Sprintf(`(?P<%s>%s[^/]+%s)`, name, prefix, suffix))
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
	return &dynamicMatcher{re: re, total: total, first: first, prefixSuffixCount: prefixSuffixCount}
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
