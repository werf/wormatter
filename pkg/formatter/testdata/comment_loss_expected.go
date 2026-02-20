package main

const (
	// SomeConst is a documented constant.
	SomeConst = "hello"

	// anotherConst is private.
	anotherConst = "world"
)

var (
	// ChartTSEntryPoints defines supported TypeScript/JavaScript entry points (in priority order).
	ChartTSEntryPoints = []string{
		"index.ts",
		"index.js",
	}
	// MaxRetries is the maximum number of retries.
	MaxRetries = 3

	// defaultTimeout is the default timeout duration.
	defaultTimeout = 30
)

func main() {}
