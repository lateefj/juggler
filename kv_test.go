package juggler

import (
	"testing"
	"time"
)

const (
	fibSize = 30
)

// Great non trivial use of CPU cycles... Well the sequence is awesome anyway
func fib(n int) int {
	if n == 1 {
		return 1
	} else if n == 0 {
		return 0
	}
	return fib(n-1) + fib(n-2)
}

func TestKVGetSetF(t *testing.T) {
	size := 100
	kv := NewKV()
	runCount := 0
	for i := range make([]int, size) {
		kv.SetF(i, func() {
			runCount++
		})
	}
	for i := range make([]int, size) {
		_, err := kv.Get(i)
		if err != nil {
			t.Errorf("Error trying to get valu %d from kv error: %s", i, err)
		}
	}

	if runCount != size {
		t.Errorf("Expected runcount to be %d but it was %d", size, runCount)
	}
	kv = NewKV()
	for i := range make([]int, size) {
		kv.SetF(i, func() {
			time.Sleep(time.Duration(i) * time.Nanosecond)
		})
	}
	for _ = range make([]int, size) {
		runCount--
		_, err := kv.Get(runCount)
		if err != nil {
			t.Errorf("Error trying to get valu %d from kv error: %s", runCount, err)
		}
	}
	if runCount != 0 {
		t.Errorf("Expected runcount to be 0 but it was %d", runCount)
	}
}
func TestKVGetSetPRF(t *testing.T) {
	// For SetPRF
	kv := NewKV()
	for i := range make([]int, fibSize) {
		kv.SetPRF(i, func(s interface{}) interface{} {
			return fib(s.(int))
		}, i)
	}
	for i := range make([]int, fibSize) {
		v, err := kv.Get(i)
		if err != nil {
			t.Errorf("Expected to have a value for %d but got error trying to get it: %s", i, err)
		}
		if v.(int) != fib(i) {
			t.Errorf("Huston we have a problem expected %d but got %d", fib(i), v)
		}
	}
}
func TestKVGetSetPF(t *testing.T) {
	// For SetPF
	kv := NewKV()
	for i := range make([]int, fibSize) {
		kv.SetPF(i, func(s interface{}) {
			fib(s.(int))
		}, i)
	}
	for i := range make([]int, fibSize) {
		_, err := kv.Get(i)
		if err != nil {
			t.Errorf("Expected to have a value for %d but got error trying to get it: %s", i, err)
		}
	}
}

func TestKVGetSetPR(t *testing.T) {
	// For SetRF
	kv := NewKV()
	for i := range make([]int, fibSize) {
		kv.SetRF(i, func() interface{} {
			return 2
		})
	}

	for i := range make([]int, fibSize) {
		v, err := kv.Get(i)
		if err != nil {
			t.Errorf("Expected to have a value for %d but got error trying to get it: %s", i, err)
		}
		if v.(int) != 2 {
			t.Errorf("Huston we have a problem expected %d but got %d", 2, v)
		}
	}
}
