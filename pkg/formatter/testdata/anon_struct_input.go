package main

import "fmt"

type Named struct {
	Zulu   string
	Alpha  string
	bravo  int
	yankee int
}

func main() {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{name: "first", input: 1, expected: "one"},
		{name: "second", input: 2, expected: "two"},
	}

	anon := struct {
		Zebra string
		Apple string
	}{Zebra: "z", Apple: "a"}

	fmt.Println(tests, anon)
}
