package prefix

import (
	"strconv"
	"testing"
)

const benchTrieKeySpace = 1 << 12

func makeBenchTrieKeys() []string {
	keys := make([]string, benchTrieKeySpace)
	for i := 0; i < benchTrieKeySpace; i++ {
		keys[i] = "user/" + strconv.Itoa(i>>8) + "/profile/" + strconv.Itoa(i)
	}
	return keys
}

func BenchmarkTriePut(b *testing.B) {
	t := NewTrie[int]()
	keys := makeBenchTrieKeys()
	mask := benchTrieKeySpace - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Put(keys[i&mask], i)
	}
}

func BenchmarkTrieGet(b *testing.B) {
	t := NewTrie[int]()
	keys := makeBenchTrieKeys()
	for i, key := range keys {
		t.Put(key, i)
	}

	mask := benchTrieKeySpace - 1
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.Get(keys[i&mask])
	}
}

func BenchmarkTrieKeysWithPrefix(b *testing.B) {
	t := NewTrie[int]()
	keys := makeBenchTrieKeys()
	for i, key := range keys {
		t.Put(key, i)
	}
	prefix := "user/7/profile/"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.KeysWithPrefix(prefix)
	}
}
