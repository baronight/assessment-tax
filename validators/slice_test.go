//go:build !integration
// +build !integration

package validators

import (
	"reflect"
	"testing"
)

func TestIsAllStringInArray(t *testing.T) {
	assertEqual := func(t *testing.T, want, got interface{}) {
		t.Helper()
		if !reflect.DeepEqual(want, got) {
			t.Errorf("expect %#v but got %#v", want, got)
		}
	}
	t.Run("given empty array b should return false", func(t *testing.T) {
		a := []string{"a", "b", "c"}
		b := []string{}

		got := IsAllStringInArray(a, b)

		assertEqual(t, false, got)
	})
	t.Run("given empty array a should return false", func(t *testing.T) {
		a := []string{}
		b := []string{"a"}

		got := IsAllStringInArray(a, b)

		assertEqual(t, false, got)
	})
	t.Run("given all value in array b had in array a should return true", func(t *testing.T) {
		a := []string{"a", "b", "c"}
		b := []string{"c", "a"}

		got := IsAllStringInArray(a, b)

		assertEqual(t, true, got)
	})
	t.Run("given some value in array b had not in array a should return false", func(t *testing.T) {
		a := []string{"a", "c"}
		b := []string{"a", "b", "c"}

		got := IsAllStringInArray(a, b)

		assertEqual(t, false, got)
	})
}
