package juggler
// All possible combination of calling a funciton types here 
// There will be lots of type assertion :(

// Start with just a funciton call
type Func func()
// Now a function that has a return value
type RFunc func() interface{}
// Now a function that takes a single data valeu
type PFunc func(data interface{})
//  Defines a type of callback basically
type PRFunc func(data interface{}) interface{}

// Value that holds some data element and a queue when that data elemetn is ready
type V struct {
	Ready bool
	Queue chan interface{}
	Data  interface{}
}

func newV() *V {
	return &V{false, make(chan interface{}, 1), nil}
}
