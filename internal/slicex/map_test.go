package slicex_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/raf555/mangoplus/internal/slicex"
)

func TestMap(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got := slicex.Map[[]int, int, string](nil, strconv.Itoa)
		if got != nil {
			t.Errorf("Map(nil) = %v, want nil", got)
		}
	})

	t.Run("empty input returns empty non-nil slice", func(t *testing.T) {
		got := slicex.Map([]int{}, strconv.Itoa)
		if got == nil {
			t.Fatal("Map([]int{}) returned nil, want non-nil empty slice")
		}
		if len(got) != 0 {
			t.Errorf("Map([]int{}) = %v, want empty slice", got)
		}
	})

	t.Run("int to string", func(t *testing.T) {
		got := slicex.Map([]int{1, 2, 3}, strconv.Itoa)
		want := []string{"1", "2", "3"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map([1,2,3], Itoa) = %v, want %v", got, want)
		}
	})

	t.Run("preserves order and length", func(t *testing.T) {
		in := []int{10, 20, 30, 40}
		got := slicex.Map(in, func(v int) int { return v * 2 })
		want := []int{20, 40, 60, 80}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map doubling = %v, want %v", got, want)
		}
	})

	t.Run("named slice type", func(t *testing.T) {
		type Ints []int
		got := slicex.Map(Ints{1, 2}, func(v int) int { return v + 1 })
		want := []int{2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map(named slice) = %v, want %v", got, want)
		}
	})
}
