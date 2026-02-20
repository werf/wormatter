package main

type tablesBuilder struct {
	logStore string

	maxLogEventTableWidth int

	discoveryClient string

	dynamicClient string
}

type KubeClient struct {
	NoActivityTimeout int

	Ownership string

	name string

	port int
}

func main() {}
