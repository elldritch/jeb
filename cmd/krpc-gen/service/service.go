package service

type Services map[string]Definition

type Definition struct {
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
	ReturnType       Type `json:"return_type"`
	ReturnIsNullable bool
	Documentation    string
}

type Parameter struct {
	Name string
	Type Type `json:"type"`
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
