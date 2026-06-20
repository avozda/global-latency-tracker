package main

import "fmt"

type Foo interface {
	Bar()
}

type FooImpl struct{}

func (f *FooImpl) Bar() {}

func getFoo() (*FooImpl, error) {
	return nil, fmt.Errorf("err")
}

func getIface() (Foo, error) {
	return getFoo()
}

func main() {
	f, _ := getIface()
	fmt.Printf("f == nil? %v\n", f == nil)
}
