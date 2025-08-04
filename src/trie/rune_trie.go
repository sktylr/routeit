package trie

// A [RuneTrie] is a structure that can be used to store strings and check for
// their presence. The lookup is case insensitive, and only characters allowed
// in a header key, per RFC-7230. These are all alphameric characters and "!",
// "#", "$", "%", "&", "'", "*", "+", "-", ".", "^", "_", "`", "|" and "~".
// Whitespace is ignored in insertion and lookup.
type RuneTrie struct {
	root *runeNode
}

type runeNode struct {
	char     string
	end      bool
	children [51]*runeNode
}

// Creates an empty [RuneTrie] that can be used to store strings and check for
// the presence of a string in the set.
func NewRuneTrie() *RuneTrie {
	return &RuneTrie{root: &runeNode{}}
}

func newRuneNode(r rune) *runeNode {
	return &runeNode{children: [51]*runeNode{}, char: string(r)}
}

// Determines whether the given input exists in the set of strings
func (t *RuneTrie) Contains(s string) bool {
	cur := t.root

	for _, r := range s {
		if r == ' ' || r == '\t' {
			continue
		}
		index := toIndex(r)
		if index == -1 || cur.children[index] == nil {
			return false
		}
		cur = cur.children[index]
	}

	return cur != nil && cur.end
}

// Inserts the given string into the trie. If the string contains illegal
// characters, or already exists in the trie, we don't insert it. Whitespace is
// removed.
func (t *RuneTrie) Insert(s string) {
	cur := t.root

	for _, r := range s {
		if r == ' ' || r == '\t' {
			continue
		}
		index := toIndex(r)
		if index == -1 {
			return
		}
		if cur.children[index] == nil {
			cur.children[index] = newRuneNode(r)
		}
		cur = cur.children[index]
	}

	if cur != nil && cur != t.root {
		cur.end = true
	}
}

// The children of each node are stored in a fixed size array of 51 elements
// (since there are 51 total allowed characters in a HTTP header key, assuming
// case insensitivity). This function will convert the character to the
// corresponding index in the children array it corresponds to. We do this to
// reduce space and only store the number of children we need (rather than 128,
// the total number of ASCII characters). Indices are ordered in the following
// way: digits, characters (case insensitive), special characters.
func toIndex(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'a' && r <= 'z':
		return 10 + int(r-'a')
	case r >= 'A' && r <= 'Z':
		return 10 + int(r-'A')
	case r == '!':
		return 36
	case r == '#' || r == '$' || r == '%' || r == '&' || r == '\'':
		return 37 + int(r-'#')
	case r == '*' || r == '+':
		return 42 + int(r-'*')
	case r == '-' || r == '.':
		return 44 + int(r-'-')
	case r == '^' || r == '_' || r == '`':
		return 46 + int(r-'^')
	case r == '|':
		return 49
	case r == '~':
		return 50
	default:
		return -1
	}
}
