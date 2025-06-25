package routeit

import "testing"

func TestTrieLookupEmpty(t *testing.T) {
	trie := newTrie[int]()

	val, found := trie.find("/foo")
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
	trie.insert("/foo", &val)

	verifyTrieElementPresent(t, trie, "/foo", val)
}

func TestTrieLookupPopulatedNotPresent(t *testing.T) {
	trie := newTrie[int]()
	trie.insert("/foo", ptr(13))
	trie.insert("/foo/bar", ptr(42))
	trie.insert("/foo/baz", ptr(19))
	trie.insert("/foo/bar/qux", ptr(10))

	actual, found := trie.find("/foo/bar/baz")
	if found {
		t.Error("did not expect to find element")
	}
	if actual != nil {
		t.Errorf("expected value to be nil")
	}
}

func TestTrieLookupNonLeafPresent(t *testing.T) {
	trie := newTrie[int]()
	trie.insert("/foo/bar/baz", ptr(42))
	trie.insert("/foo/baz", ptr(19))
	trie.insert("/foo/bar/qux", ptr(10))

	actual, found := trie.find("/foo/bar")
	if found {
		t.Error("did not expect to find element")
	}
	if actual != nil {
		t.Errorf("expected value to be nil")
	}
}

func TestTrieLookupForcesLeadingSlash(t *testing.T) {
	val := 56
	trie := newTrie[int]()
	trie.insert("/foo/bar/baz", &val)

	verifyTrieElementPresent(t, trie, "foo/bar/baz", val)
	verifyTrieElementPresent(t, trie, "/foo/bar/baz", val)
}

func TestTrieLookupIgnoresTrailingSlash(t *testing.T) {
	val := 17
	trie := newTrie[int]()
	trie.insert("/foo/bar/baz", &val)

	verifyTrieElementPresent(t, trie, "/foo/bar/baz/", val)
	verifyTrieElementPresent(t, trie, "/foo/bar/baz", val)
}

func TestTrieInsertForcesLeadingSlash(t *testing.T) {
	val := 45
	trie := newTrie[int]()
	trie.insert("foo/bar", &val)

	verifyTrieElementPresent(t, trie, "/foo/bar", val)
}

func TestTrieInsertIgnoresTrailingSlash(t *testing.T) {
	val := 18
	trie := newTrie[int]()
	trie.insert("/foo/bar/", &val)

	verifyTrieElementPresent(t, trie, "/foo/bar", val)
}

func verifyTrieElementPresent(t *testing.T, trie *trie[int], key string, want int) {
	t.Helper()
	actual, found := trie.find(key)
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
