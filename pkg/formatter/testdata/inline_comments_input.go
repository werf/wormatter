package main

const (
	ZetaConst  = "zeta"  // zeta inline comment
	AlphaConst = "alpha" // alpha inline comment
	BetaConst  = "beta"  // beta inline comment
)

const (
	// doc comment for TypedZ
	TypedZ MyType = "z"
	// doc comment for TypedA
	TypedA MyType = "a" // typed-a inline
)

const SingleWithComment = "single" // single inline comment
const AnotherSingle = "another"    // another inline comment

var (
	varZ = 1 // var-z inline
	varA = 2 // var-a inline
)

type MyType string

func main() {}
