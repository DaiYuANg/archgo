package prefix

import (
	"encoding/json"

	common "github.com/DaiYuANg/arcgo/collectionx/internal"
	"github.com/samber/lo"
)

// All returns all key-value pairs as a copied map.
func (t *Trie[V]) All() map[string]V {
	pairs := t.pairsWithPrefix("")
	if len(pairs) == 0 {
		return map[string]V{}
	}

	out := make(map[string]V, len(pairs))
	lo.ForEach(pairs, func(item keyValue[V], _ int) {
		out[item.key] = item.value
	})
	return out
}

// ToJSON serializes all key-value pairs to JSON.
func (t *Trie[V]) ToJSON() ([]byte, error) {
	return json.Marshal(t.All())
}

// MarshalJSON implements json.Marshaler.
func (t *Trie[V]) MarshalJSON() ([]byte, error) {
	return t.ToJSON()
}

// String implements fmt.Stringer.
func (t *Trie[V]) String() string {
	data, err := t.ToJSON()
	return common.JSONResultString(data, err, "{}")
}
