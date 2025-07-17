// This file contains the types and functions used to model a trie. Unlike a
// regular trie which splits on each character, this trie splits on the '/'
// character and is used to model a URL path hierarchy. The trie only supports
// inserts and lookups and currently does not support pattern matches - only
// direct matches.
//
// https://www.geeksforgeeks.org/dsa/trie-insert-and-search/

package routeit

import (
	"errors"
	"regexp"
	"strings"
)

type trie[T any] struct {
	root *node[T]
}

// Trie keys can match exactly, or dynamically against the input key. This
// struct allows the trie to keep track of the kind of key, to ensure that
// insertion and lookup obeys the concept.
type trieKey struct {
	exact string
	// TODO: look into just removing this?
	wildcard bool
}

// A dynamic matcher is used in value nodes to signify that there is at least
// one component of that node's path that is dynamic in nature. This stores the
// regex for the path, which is a named character matcher, and also the total
// number of dynamic components and the position of the first occurrence of a
// dynamic component in the path, which are both used for prioritisation.
type dynamicMatcher struct {
	re    *regexp.Regexp
	total int
	first int
}

type trieValue[T any] struct {
	dm  *dynamicMatcher
	val *T
}

type node[T any] struct {
	key      trieKey
	value    *trieValue[T]
	children []*node[T]
}

func newKey(part string) trieKey {
	if strings.HasPrefix(part, ":") {
		// This is a wildcard matcher and will match against anything
		return trieKey{wildcard: true}
	}
	return trieKey{exact: part}
}

func newTrie[T any]() *trie[T] {
	return &trie[T]{root: &node[T]{}}
}

// TODO: can probably move this to its own package
func (t *trie[T]) Find(path string) (*T, pathParameters, bool) {
	if t.root == nil {
		return nil, pathParameters{}, false
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
			return nil, pathParameters{}, false
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
		return nil, pathParameters{}, false
	}

	// We omit the nil check on the inner value since by construction it should
	// always be populated.
	return found.value.val, found.value.PathParams(path), true
}

func (t *trie[T]) Insert(path string, value *T) {
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

	dynamicMatcher := dynamicPathToMatcher(path)
	if dynamicMatcher == nil {
		current.value = &trieValue[T]{val: value}
		return
	}

	if current.value != nil && current.value.dm.re.String() != dynamicMatcher.re.String() {
		panic(errors.New("multiple dynamic handlers registered to the same route"))
	}

	current.value = &trieValue[T]{val: value, dm: dynamicMatcher}
}

func (n *node[T]) GetOrCreateChild(key string) *node[T] {
	for _, child := range n.children {
		if child.key.exact == key {
			// We don't use the wildcard comparison here, otherwise we would
			// match all static paths against dynamic paths, causing some nodes
			// to be overwritten depending on the order of insertions.
			return child
		}
	}
	newChild := &node[T]{key: newKey(key)}
	n.children = append(n.children, newChild)
	return newChild
}

func (k *trieKey) Matches(cmp string) bool {
	if k.wildcard {
		return true
	}
	return k.exact == cmp
}

// Determines whether a node has strictly higher priority than another node. If
// n is a static node (i.e. no parts of its path are dynamic), then it has
// higher priority than anything else. If n is not static and other is, then
// other takes priority. If both are dynamic, then we compare their dynamic
// components. If n has strictly less dynamic components than other, n takes
// priority. If they have the same, we compare the specificity of the dynamic
// components. Dynamic components that appear earlier in the path are less
// specific.
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
	if n.value.dm.total < other.value.dm.total {
		return true
	}
	if n.value.dm.total == other.value.dm.total {
		return n.value.dm.first > other.value.dm.first
	}
	return false
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

// Constructs a dynamic matcher for a given path, returning nil if the path has
// no dynamic components. This includes building a named regex that can be used
// to extract the path parameters of the request once matched.
func dynamicPathToMatcher(path string) *dynamicMatcher {
	if !strings.Contains(path, ":") {
		return nil
	}

	// TODO: some of the leading slash stuff makes this more confusing than it should be

	first, total := int(^uint(0)>>1), 0
	var sb strings.Builder
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
			// We have a segment that is ":name". We want to convert this into
			// a named character group that allows ASCII characters, stopping
			// at the first /.
			sb.WriteString("(?P<")
			sb.WriteString(seg[1:])
			sb.WriteString(">[^/]+)")
			total++
			if i < first {
				first = i
			}
		}
	}

	re := regexp.MustCompile(sb.String())
	return &dynamicMatcher{re: re, total: total, first: first}
}
