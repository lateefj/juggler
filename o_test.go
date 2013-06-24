package juggler

import (
	"testing"
)

func TestOrderedF(t *testing.T) {

	size := 100
	o := NewO()
	count := 0
	for i := 0; i < size; i++ {
		o.AddF(func() {
		})
	}
	for _ = range o.Range() {
		count++
	}
	if count != size {
		t.Errorf("Expected count to be 0 but it was %d", count)
	}
}
func TestOrderedRF(t *testing.T) {

	size := 30
	o := NewO()
	count := 0
	for i := 0; i < size; i++ {
		o.AddRF(func() interface{} {
			return fib(size)
		})
	}

	for v := range o.Range() {
		if v != fib(size) {
			t.Errorf("Expect value to be %d but was %d", fib(size), v)
		}
		count++
	}
}
func TestOrderedPF(t *testing.T) {

	size := 30
	o := NewO()
	count := 0
	for i := 0; i < size; i++ {
		o.AddPF(func(s interface{}) {
			if s.(int) < 0 || s.(int) > size {
				t.Errorf("Expected s to be between 0 and %d but was %d", size, s.(int))
			}
		}, i)
	}

	for _ = range o.Range() {
		count++
	}
	if count != size {
		t.Errorf("Expected count to be 0 but it was %d", count)
	}
}

func TestOrderedPRF(t *testing.T) {
	size := 30
	o := NewO()
	count := 0
	for i := 0; i < size; i++ {
		o.AddPRF(func(data interface{}) interface{} {
			return fib(data.(int))
		}, i)
	}
	for v := range o.Range() {
		if v != fib(count) {
			t.Errorf("Expect value to be %d but was %d", fib(size), v)
		}
		count++
	}
}
