package main

import "fmt"

var (
	FeatGates = []*FeatGate{}
	Alpha     = NewFeatGate("alpha")
	Beta      = NewFeatGate("beta")
	Gamma     = NewFeatGate("gamma")

	privateFirst  = 1
	privateSecond = 2
)

type FeatGate struct {
	Name string
}

func NewFeatGate(name string) *FeatGate {
	fg := &FeatGate{Name: name}
	FeatGates = append(FeatGates, fg)

	return fg
}

func main() {
	fmt.Println(FeatGates)
}
