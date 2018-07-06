package main

import "flag"

func main() {
	_ = flag.String("dir", "", "directory of JSON service definitions")
	flag.Parse()
}
