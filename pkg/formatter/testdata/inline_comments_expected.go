package main

const (
	// doc comment for TypedZ
	TypedZ MyType = "z"
	// doc comment for TypedA
	TypedA MyType = "a" // typed-a inline

	AlphaConst        = "alpha"   // alpha inline comment
	AnotherSingle     = "another" // another inline comment
	BetaConst         = "beta"    // beta inline comment
	SingleWithComment = "single"  // single inline comment
	// zeta inline comment
	ZetaConst = "zeta"
)

var (
	varZ = 1 // var-z inline
	// var-a inline
	varA = 2
)

type MyType string

func main() {}
