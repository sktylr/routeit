package requestid

import (
	"regexp"
	"slices"
	"testing"

	"github.com/sktylr/routeit"
)

func TestUuidV7(t *testing.T) {
	prov := NewUuidV7Provider()

	t.Run("generates UUID", func(t *testing.T) {
		regex := regexp.MustCompile(
			"[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}",
		)

		id := prov(&routeit.Request{})

		if !regex.MatchString(id) {
			t.Errorf(`id = %#q, does not match UUID regex`, id)
		}
	})

	t.Run("guarantees order by creation", func(t *testing.T) {
		id1 := prov(&routeit.Request{})
		id2 := prov(&routeit.Request{})

		if id1 == id2 {
			t.Errorf(`id1 = %#q, id2 = %#q, should not equal`, id1, id2)
		}
		if !slices.IsSorted([]string{id1, id2}) {
			t.Errorf(`id1 = %#q, id2 = %#q, not in order`, id1, id2)
		}
	})
}
