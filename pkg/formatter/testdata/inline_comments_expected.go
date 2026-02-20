package main

const (
	AlphaConst        = "alpha"   // alpha inline comment
	AnotherSingle     = "another" // another inline comment
	BetaConst         = "beta"    // beta inline comment
	SingleWithComment = "single"  // single inline comment
	ZetaConst         = "zeta"    // zeta inline comment

	// doc comment for TypedA
	TypedA MyType = "a" // typed-a inline
	// doc comment for TypedZ
	TypedZ MyType = "z"
)

var (
	varZ = 1 // var-z inline
	// var-a inline
	varA = 2
)

type MyType string

func main() {}
