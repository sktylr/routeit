## Trie

Routing and URL rewriting both use the same data structure, which is modelled off a trie.
A trie is a tree-like structure used to store strings in dictionaries or sets, and is often referred to as a prefix tree or radix tree.
Each nodes stores a single character, which allows for fast search for keys.

The trie-like structure used in this project is similar, but has its differences.
Firstly, the separator used here is a path separator - `/`.
So each node stores a string.

### Nodes

All nodes in the trie have a key, and some will have a value.
If after traversing the trie, the resultant node has a value, then the corresponding key is in the trie and can be returned to the caller.

The trie supports both dynamic and static lookup of keys.
Additionally, dynamic lookups can be further controlled by requiring specific prefixes or suffixes on the dynamically matched path component.

### Syntax

A static path component is written using the exact string it should match against.
For example, `/foo/bar` matches exactly against the string `"/foo/bar"` and nothing else.
Dynamic path component use the colon (`:`) character to indicate the path component is dynamic.
The colon must be the first character after the `/`, i.e. it must come at the start of the path component.

The characters after the colon, up to the next slash or pipe (`|`), construct the name used to refer to the matched component.
So `/:foo` matches against `/<anything>`, and assigns the name `foo` to all matched characters after the leading slash.
Note this will only match against single path component paths (i.e. `/foo` but not `/foo/bar`).

Required prefixes or suffixes are indicated using the pipe character (`|`) after the path component's name.
The prefix is specified after the first pipe, while the suffix is after the second.
A dynamic component may contain 0, 1 or 2 pipes and there does not need to be any characters between the pipes.
For example, `/:foo|prefix`, `/:foo||suffix`, `/:foo||` and `/:foo|` are all valid syntaxes for paths featuring dynamic components.
The first one requires a prefix of `prefix` on the first path component, while the second requires a suffix of `suffix` on the first path component.
The remaining two are both functionally equivalent to `/:foo` (match against anything with no required prefixes nor suffixes).

### Specificity

Due to static, dynamic and dynamic with prefixes and/or suffixes, a key can conceivably match against multiple nodes.
For example, given a trie containing the keys `/prefixes`, `/:foo`, `/:foo|prefix` and `/:foo||es`, the input `/prefixes` will match against all keys presented.
The trie needs a reliable way to determine which value to select, and ideally be deterministic in which values are selected, which is where path specificity comes in.

The trie calculates the specificity of each of the ultimately matched nodes in four phases after performing BFS.
If the nodes still cannot be separated after the fourth phase, we then take whichever appeared first in the nodes traversed, which corresponds to the order of insertion.

#### Phase 1: Static Components

A node whose path contains exactly 0 dynamic components is the most specific.
This is due to the trie construction - insertion prevents conflicting insertions, meaning we are guaranteed to only ever have at most 1 completely static node that matches the incoming key.

#### Phase 2: Number of Dynamic Components

When comparing two nodes that contain dynamic matches in their path, we first assess the number of dynamic path components in each node's path.
A node A that has strictly less dynamic path components that another node B is strictly more specific than B, meaning it takes precedence.
Technically this is equivalent to Phase 1 above, but is more explicit in the metric of specificity.

#### Phase 3: Number of Prefixes and Suffixes

If the nodes still cannot be separated, we count the total number of required prefixes and suffixes on both paths.
Since the input key matched against both nodes, we know the key matched against the required prefixes and suffixes.
A node A is strictly more specific (given inseparable specificity in Phase 2) than another node B if A has strictly more prefixes and suffixes in its dynamic path components compositions.

#### Phase 4: Leading Static Components

If we still cannot separate the nodes, we locate the first occurrence of a dynamic path component in their respective paths.
For example, `/foo/bar/:baz`'s first appearance is in position 2, assuming 0-indexing.
A node A is strictly more specific than another node B (given inseparable specificity in Phase 3) if A's earliest appearance of a dynamic path component is strictly later than that of B.
This is because the number of leading static components in A is higher, meaning we are deeper into the trie before we reach a dynamic component.

An example of the measures of specificity is shown below.

| A                       | B                | Comparing       | More specific | Phase | Reason                         |
| ----------------------- | ---------------- | --------------- | ------------- | ----- | ------------------------------ |
| `/foo/bar`              | `/foo/:baz`      | `/foo/bar`      | A             | 1     | Static path                    |
| `/:foo/bar`             | `/foo/:bar`      | `/foo/bar`      | B             | 4     | More leading static components |
| `/foo/:bar/baz`         | `/:foo/bar/:baz` | `/foo/bar/baz`  | A             | 2     | Less dynamic components        |
| `/foo/:bar`             | `/foo/:bar\|baz` | `/foo/baza`     | B             | 3     | More prefixes                  |
| `/foo/:bar/:baz\|\|qux` | `/foo/:bar/:baz` | `/foo/bar/aqux` | A             | 3     | More suffixes                  |

> [!NOTE]
> Due to how prefixes and suffixes are chosen, there can be ambiguity with separate routes that match against the same path space that use prefixes and suffixes.
> For example, the routes `/foo/:bar|baz` and `/foo/:bar||qux` could both match against the same input (e.g. `/foo/bazqux`).
> Currently, `routeit` does not provide higher precedence to individual prefixes or suffixes, so these cannot be separated.
> It is best to avoid these types of matches and only use dynamic routes with prefixes and suffixes sparingly.

### Extraction

Once a key matches against a node that contains at least 1 dynamic path component, the dynamic components need to be extracted.
At insertion time, a regex is compiled that allows for easy extraction.
For example, for URL routing, we would want to extract the named variables of the path that were dynamically matched against, as they typically represent useful information such as an ID.

Where a dynamic component requires prefixes or suffixes, the prefix or suffix will **not** be stripped from the matched component.
For example, given the path `/:foo|prefix` and the lookup key `/prefixed`, the entire string `"prefixed"` would be assigned to the `foo` variable, not just `ed` (the characters after the required prefix).
