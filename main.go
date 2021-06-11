package main

import "fmt"

func main() {

	g := createTestGraph()
	r := g.DagRun()

	r.Run()
	for res := range r.errors {
		fmt.Println(res)
	}
}

