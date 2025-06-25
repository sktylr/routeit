// This file contains the types and functions used to model a trie. Unlike a
// regular trie which splits on each character, this trie splits on the '/'
// character and is used to model a URL path hierarchy. The trie only supports
// inserts and lookups and currently does not support pattern matches - only
// direct matches.
//
// https://www.geeksforgeeks.org/dsa/trie-insert-and-search/

package routeit

import (
	"strings"
)

type trie[T any] struct {
	root *node[T]
}

type node[T any] struct {
	key      string
	value    *T
	children []*node[T]
}

func newTrie[T any]() *trie[T] {
	return &trie[T]{root: &node[T]{}}
}

// TODO: can probably move this to its own package
func (t *trie[T]) find(path string) (*T, bool) {
	if t.root == nil {
		return nil, false
	}

	current := t.root
	for seg := range strings.SplitSeq(path, "/") {
		if seg == "" {
			continue
		}
		found := false
		for _, child := range current.children {
			if child.key == seg {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
	}

	return current.value, current.value != nil
}

func (t *trie[T]) insert(path string, value *T) {
	if t.root == nil {
		t.root = &node[T]{}
	}

	current := t.root
	for seg := range strings.SplitSeq(path, "/") {
		if seg == "" {
			continue
		}
		current = current.getOrCreateChild(seg)
	}

	current.value = value
}

func (n *node[T]) getOrCreateChild(key string) *node[T] {
	for _, child := range n.children {
		if child.key == key {
			return child
		}
	}
	newChild := &node[T]{key: key}
	n.children = append(n.children, newChild)
	return newChild
}
