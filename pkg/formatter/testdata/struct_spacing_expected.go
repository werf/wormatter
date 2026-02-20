package main

type tablesBuilder struct {
	discoveryClient       string
	dynamicClient         string
	logStore              string
	maxLogEventTableWidth int
}

type KubeClient struct {
	NoActivityTimeout int
	Ownership         string

	name string
	port int
}

func main() {}
