package types

import (
	"reflect"
	"strconv"
	"testing"
)

func TestAddDuplicateItem(t *testing.T) {
	set := NewSet()
	set.Add(`test`)
	set.Add(`test`)

	if !reflect.DeepEqual([]interface{}{`test`}, set.Flatten()) {
		t.Errorf(`Incorrect result returned: %+v`, set.Flatten())
	}
}

func TestAddItems(t *testing.T) {
	set := NewSet()
	set.Add(`test`)
	set.Add(`test1`)

	firstSeen := false
	secondSeen := false
	// order is not guaranteed
	for _, item := range set.Flatten() {
		if item.(string) == `test` {
			firstSeen = true
		} else if item.(string) == `test1` {
			secondSeen = true
		}
	}

	if !firstSeen || !secondSeen {
		t.Errorf(`Not all items seen in set.`)
	}
}

func TestRemove(t *testing.T) {
	set := NewSet()
	set.Add(`test`)
	set.Remove(`test`)

	if !reflect.DeepEqual([]interface{}{}, set.Flatten()) {
		t.Errorf(`Incorrect result returned: %+v`, set.Flatten())
	}
}

func TestExists(t *testing.T) {
	set := NewSet()
	set.Add(`test`)

	if !set.Exists(`test`) {
		t.Errorf(`Correct existence not determined`)
	}

	if set.Exists(`test1`) {
		t.Errorf(`Correct nonexistence not determined.`)
	}
}

func TestLen(t *testing.T) {
	set := NewSet()
	set.Add(`test`)

	if set.Len() != 1 {
		t.Errorf(`Expected len: %d, received: %d`, 1, set.Len())
	}

	set.Add(`test1`)
	if set.Len() != 2 {
		t.Errorf(`Expected len: %d, received: %d`, 2, set.Len())
	}
}

func TestFlattenCaches(t *testing.T) {
	set := NewSet()
	item := `test`
	set.Add(item)

	set.Flatten()

	if len(set.flattened) != 1 {
		t.Errorf(`Expected len: %d, received: %d`, 1, len(set.flattened))
	}
}

func TestAddClearsCache(t *testing.T) {
	set := NewSet()
	item := `test`
	set.Add(item)
	set.Flatten()

	set.Add(item)

	if len(set.flattened) != 0 {
		t.Errorf(`Expected len: %d, received: %d`, 0, len(set.flattened))
	}

	item = `test2`
	set.Add(item)

	if set.flattened != nil {
		t.Errorf(`Cache not cleared.`)
	}
}

func TestDeleteClearsCache(t *testing.T) {
	set := NewSet()
	item := `test`
	set.Add(item)
	set.Flatten()

	set.Remove(item)

	if set.flattened != nil {
		t.Errorf(`Cache not cleared.`)
	}
}

func TestAll(t *testing.T) {
	set := NewSet()
	item := `test`
	set.Add(item)

	result := set.All(item)
	if !result {
		t.Errorf(`Expected true.`)
	}

	itemTwo := `test1`

	result = set.All(item, itemTwo)
	if result {
		t.Errorf(`Expected false.`)
	}
}

func TestClear(t *testing.T) {
	set := NewSet()
	set.Add(`test`)

	set.Clear()

	if set.Len() != 0 {
		t.Errorf(`Expected len: %d, received: %d`, 0, set.Len())
	}
}

func BenchmarkFlatten(b *testing.B) {
	set := NewSet()
	for i := 0; i < 50; i++ {
		item := strconv.Itoa(i)
		set.Add(item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Flatten()
	}
}

func BenchmarkLen(b *testing.B) {
	set := NewSet()
	for i := 0; i < 50; i++ {
		item := strconv.Itoa(i)
		set.Add(item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Len()
	}
}

func BenchmarkExists(b *testing.B) {
	set := NewSet()
	set.Add(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Exists(1)
	}
}

func BenchmarkClear(b *testing.B) {
	set := NewSet()
	for i := 0; i < b.N; i++ {
		set.Clear()
	}
}
