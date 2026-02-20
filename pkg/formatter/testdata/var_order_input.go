package main

import "fmt"

var FeatGates = []*FeatGate{}

func NewFeatGate(name string) *FeatGate {
	fg := &FeatGate{Name: name}
	FeatGates = append(FeatGates, fg)

	return fg
}

var Alpha = NewFeatGate("alpha")
var Beta = NewFeatGate("beta")
var Gamma = NewFeatGate("gamma")

var privateFirst = 1
var privateSecond = 2

type FeatGate struct {
	Name string
}

func main() {
	fmt.Println(FeatGates)
}
