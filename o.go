package juggler

// Maintains the order of when the functions are added
type O struct {
	kv   *KV
	size int // might need to do some sync stuff here?
}

func (o *O) Range() chan interface{} {
	l := make(chan interface{})
	go func() { // Push them all in order on a channel :)
		for i := 0; i < o.size; i++ {
			v, err := o.kv.Get(i)
			if err != nil { // Something really bad happens if we panic
				panic(err)
			}
			l <- v
		}
		close(l)
	}()
	return l
}

// For functions that don't take any parameters or return anything
func (o *O) AddF(f Func) {
	o.kv.SetF(o.size, f)
	o.size++
}
// For functions that just return a value but do not take any parameters
func (o *O) AddRF(rf RFunc) {
	o.kv.SetRF(o.size, rf)
	o.size++
}
// For function that take a value but do not return anything
func (o *O) AddPF(pf PFunc, data interface{}) {
	o.kv.SetPF(o.size, pf, data)
	o.size++
}
// For functions that take a parameter and return a value (maybe should be default?)
func (o *O) AddPRF(prf PRFunc, data interface{}) {
	o.kv.SetPRF(o.size, prf, data)
	o.size++
}
func NewO() *O {
	return &O{NewKV(), 0}
}
