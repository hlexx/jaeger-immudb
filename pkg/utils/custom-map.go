package utils

import (
	"bytes"
	"sort"
)

type UnsortedPair struct {
	key   string
	value interface{}
}

// Key UnsortedPair.
func (kv *UnsortedPair) Key() string {
	return kv.key
}

// Value UnsortedPair.
func (kv *UnsortedPair) Value() interface{} {
	return kv.value
}

type SortedPair struct {
	Pairs    []*UnsortedPair
	LessFunc func(a *UnsortedPair, j *UnsortedPair) bool
}

// Len SortedPair.
func (a SortedPair) Len() int { return len(a.Pairs) }

// Swap SortedPair.
func (a SortedPair) Swap(i, j int) { a.Pairs[i], a.Pairs[j] = a.Pairs[j], a.Pairs[i] }

// Less SortedPair.
func (a SortedPair) Less(i, j int) bool { return a.LessFunc(a.Pairs[i], a.Pairs[j]) }

type SortedMap struct {
	keys       []string
	buffer     bytes.Buffer
	values     map[string]interface{}
	escapeHTML bool
}

// NewMap SortedMap.
func NewMap() *SortedMap {
	result := &SortedMap{}
	result.keys = []string{}
	result.values = map[string]interface{}{}
	result.escapeHTML = true
	result.buffer = bytes.Buffer{}
	return result
}

// Get value SortedMap.
func (uMap *SortedMap) Get(key string) (val interface{}, exists bool) {
	val, exists = uMap.values[key]
	return
}

// Set value SortedMap.
func (uMap *SortedMap) Set(key string, value interface{}) {
	_, exists := uMap.values[key]
	if !exists {
		uMap.keys = append(uMap.keys, key)
	}
	uMap.values[key] = value
}

// Keys SortedMap.
func (uMap *SortedMap) Keys() []string {
	return uMap.keys
}

// Remove value SortedMap.
func (uMap *SortedMap) Remove(key string) {
	_, exists := uMap.values[key]
	if !exists {
		return
	}
	for i, k := range uMap.keys {
		if k == key {
			uMap.keys = append(uMap.keys[:i], uMap.keys[i+1:]...)
			break
		}
	}
	delete(uMap.values, key)
}

// SortKeys map keys using your sort func
func (uMap *SortedMap) SortKeys(sortFunc func(keys []string)) {
	sortFunc(uMap.keys)
}

// SortPairs using your sort func
func (uMap *SortedMap) SortPairs(lessFunc func(a *UnsortedPair, b *UnsortedPair) bool) {
	pairs := make([]*UnsortedPair, len(uMap.keys))
	for i, key := range uMap.keys {
		pairs[i] = &UnsortedPair{key, uMap.values[key]}
	}
	sort.Sort(SortedPair{pairs, lessFunc})
	for i, pair := range pairs {
		uMap.keys[i] = pair.key
	}
}
