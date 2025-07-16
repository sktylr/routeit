package routeit

import "testing"

func TestTrieLookupEmpty(t *testing.T) {
	trie := newTrie[int]()

	val, found := trie.Find("/foo")
	if found {
		t.Error("did not expect to find element")
	}
	if val != nil {
		t.Errorf("expected value to be nil")
	}
}

func TestTrieLookupOneElement(t *testing.T) {
	val := 13
	trie := newTrie[int]()
	trie.Insert("/foo", &val)

	verifyTrieElementPresent(t, trie, "/foo", val)
}

func TestTrieLookupPopulatedNotPresent(t *testing.T) {
	trie := newTrie[int]()
	trie.Insert("/foo", ptr(13))
	trie.Insert("/foo/bar", ptr(42))
	trie.Insert("/foo/baz", ptr(19))
	trie.Insert("/foo/bar/qux", ptr(10))

	actual, found := trie.Find("/foo/bar/baz")
	if found {
		t.Error("did not expect to find element")
	}
	if actual != nil {
		t.Errorf("expected value to be nil")
	}
}

func TestTrieLookupNonLeafPresent(t *testing.T) {
	trie := newTrie[int]()
	trie.Insert("/foo/bar/baz", ptr(42))
	trie.Insert("/foo/baz", ptr(19))
	trie.Insert("/foo/bar/qux", ptr(10))

	actual, found := trie.Find("/foo/bar")
	if found {
		t.Error("did not expect to find element")
	}
	if actual != nil {
		t.Errorf("expected value to be nil")
	}
}

func verifyTrieElementPresent(t *testing.T, trie *trie[int], key string, want int) {
	t.Helper()
	actual, found := trie.Find(key)
	if !found {
		t.Error("expected to find element")
	}
	if *actual != want {
		t.Errorf(`trie["%s"] = %d, wanted %d`, key, actual, want)
	}
}

func ptr[T any](v T) *T {
	return &v
}
