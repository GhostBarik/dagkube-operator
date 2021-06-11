package main

import "fmt"

func main() {

	g := createTestGraph()
	r := g.runGraph()

	r.run()
	for res := range r.errors {
		fmt.Println(res)
	}
}

