package juggler

import (
	"fmt"
	"errors"
)

// As much as reading Generics in Java or Dart makes me nocuous it would be really helpful right. It would be nice not to have to implement every possible key type.
type KV struct {
	M map[interface{}]*V
}

// Get 
func (kv *KV) Get(k interface{}) (interface{}, error) {
	// Make sure the k is in the list of keys
	if v, ok := kv.M[k]; ok {
		// If the data is not ready then wait until it is
		if !v.Ready {
			v.Data = <-v.Queue
			v.Ready = true
		}
		return v.Data, nil
	}
	// Return an error if the key doesn't exist
	return nil, errors.New(fmt.Sprintf("Key does '%s' not exist!\n", k))
}

// Abstracts creating a value and setting it in the map
func (kv *KV) set(k interface{}) *V {
	v := newV()
	kv.M[k] = v
	return v
}

// For functions that don't take any parameters or return anything
func (kv *KV) SetF(k interface{}, f Func) {
	v := kv.set(k)
	go func() {
		f()
		v.Queue <- nil
	}()
}
// For functions that just return a value but do not take any parameters
func (kv *KV) SetRF(k interface{}, rf RFunc) {
	v := kv.set(k)
	go func() {
		d := rf()
		v.Queue <- d
	}()
}
// For function that take a value but do not return anything
func (kv *KV) SetPF(k interface{}, pf PFunc, data interface{}) {
	v := kv.set(k)
	go func() {
		pf(data)
		v.Queue <- nil
	}()
}
// For functions that take a parameter and return a value (maybe should be default?)
func (kv *KV) SetPRF(k interface{}, prf PRFunc, data interface{}) {
	v := kv.set(k)
	go func() {
		d := prf(data)
		v.Queue <- d
	}()
}

func NewKV() *KV {
	return &KV{make(map[interface{}]*V)}
}
