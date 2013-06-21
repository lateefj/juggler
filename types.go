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

// Abstract way to handle when a function finishes running
type value struct {
	Ready bool
	Data  interface{}
}

type V struct {
	value
	Queue chan interface{}
}

func newV() *V {
	return &V{value{false, nil}, make(chan interface{}, 1)}
}

type IntV struct {
	value
	Queue chan int
}

func newIntV() *IntV {
	return &IntV{value{false, nil}, make(chan int, 1)}
}

type StringV struct {
	value
	Queue chan string
}

func newStringV() *StringV {
	return &StringV{value{false, nil}, make(chan string, 1)}
}
