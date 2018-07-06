package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/dave/jennifer/jen"
)

type Service map[string]ServiceDefinition

type ServiceDefinition struct {
	ID            int
	Documentation string
	Procedures    map[string]Procedure
	Classes       map[string]Class
	Enumerations  map[string]Enumeration
	Exceptions    map[string]struct{}
}

type Procedure struct {
	ID               int
	Parameters       []Parameter
	ReturnType       Type
	ReturnIsNullable bool
	Documentation    string
}

type Parameter struct {
	Name string
	Type Type
}

type Type struct {
	Code    string
	Service string
	Name    string
	Types   []Type
}

type Class struct {
	Documentation string
}

type Enumeration struct {
	Documentation string
	Values        []Value
}

type Value struct {
	Name          string
	Value         int
	Documentation string
}

func main() {
	dir := flag.String("dir", "", "directory of JSON service definitions")
	flag.Parse()

	ls, err := ioutil.ReadDir(*dir)
	if err != nil {
		panic(err)
	}

	for _, file := range ls {
		contents, err := ioutil.ReadFile(filepath.Join(*dir, file.Name()))
		if err != nil {
			panic(err)
		}

		var service Service
		err = json.Unmarshal(contents, &service)
		if err != nil {
			panic(err)
		}

		log.Printf("%#v\n", service)
	}

	jen.NewFile("krpc")
}
