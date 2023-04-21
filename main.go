package main

import "flag"

func main() {
	path := flag.String("server", "", "Server path")
	flag.Parse()

	if path == nil || *path == "" {
		cmdServer()
	} else {
		cmdClient(*path)
	}
}
