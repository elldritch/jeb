package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
)

func main() {
	// Parse flags.
	dir := flag.String("dir", "", "directory of JSON service definitions (required)")
	flag.Parse()

	// Validate flags.
	if *dir == "" {
		flag.Usage()
		os.Exit(1)
	}

	// List service definitions.
	ls, err := ioutil.ReadDir(*dir)
	if err != nil {
		panic(err)
	}

	for _, file := range ls {
		// Read service definition.
		contents, err := ioutil.ReadFile(filepath.Join(*dir, file.Name()))
		if err != nil {
			panic(err)
		}

		// Parse service definition.
		var service Service
		err = json.Unmarshal(contents, &service)
		if err != nil {
			panic(err)
		}

		// Generate service client.
		for name, definition := range service {
			// Compute method tables.
			//
			// > Procedures names are CamelCase. Whether a procedure is a service
			// > procedure, class method, class property, and what class (if any) it
			// > belongs to is determined by its name:
			// >
			// > * `ProcedureName` - a standard procedure that is just part of a
			//     service.
			// > * `get_PropertyName` - a procedure that returns the value of a
			//     property in a service.
			// > * `set_PropertyName` - a procedure that sets the value of a property
			//     in a service.
			// > * `ClassName_MethodName` - a class method.
			// > * `ClassName_static_StaticMethodName` - a static class method.
			// > * `ClassName_get_PropertyName` - a class property getter.
			// > * `ClassName_set_PropertyName` - a class property setter.
			// >
			// > Only letters and numbers are permitted in class, method and property
			// > names. Underscores can therefore be used to split the name into its
			// > constituent parts.
			classes := make(map[string]bool)
			for class := range definition.Classes {
				classes[class] = true
			}
			statics := make(map[string]map[string]Procedure)
			methods := make(map[string]map[string]Procedure)
			for name, proc := range definition.Procedures {
				// Parse procedure name.
				splits := strings.Split(name, "_")
				if len(splits) < 2 {
					continue
				}
				class := splits[0]
				_, ok := classes[class]
				if !ok {
					continue
				}
				// Select table and ensure it's non-nil.
				var table map[string]Procedure
				if splits[1] == "static" {
					table = statics[class]
					if table == nil {
						statics[class] = make(map[string]Procedure)
						table = statics[class]
					}
				} else {
					table = methods[class]
					if table == nil {
						methods[class] = make(map[string]Procedure)
						table = methods[class]
					}
				}
				// Set procedure.
				table[name] = proc
			}

			// New file.
			file := jen.NewFile("krpc")

			// Service singleton.
			file.Type().Id(name).Struct()

			// Classes.
			for class := range definition.Classes {
				// Define class struct.
				file.Type().Id(class).Struct(
					jen.Id("id").Uint64(),
				)

				// Define static class struct and methods.
				if len(statics[class]) > 0 {
					file.Type().Id(class + "Static").Struct()
					for name, proc := range statics[class] {
						// Take idiomatic parameters
						var params []jen.Code
						for _, param := range proc.Parameters {
							params = append(params, JenType(param.Type, jen.Id(param.Name)))
						}

						static := file.Func().Params(
							jen.Id("_").Id("*" + class + "Static"),
						).Id(strings.TrimPrefix(name, class+"_static_")).Params(params...)

						if proc.ReturnType.Code != "" {
							JenType(proc.ReturnType, static)
						}

						static.Block()
					}
				}

				// Define instance methods.
				// TODO: need to handle: first param named "this"
				// Define getters.
				// Define setters.
			}

			// Enumerations.
			for enum, values := range definition.Enumerations {
				// Declare enumeration type.
				file.Type().Id(enum).Int()

				// Define enumeration values.
				firstValue := values.Values[0]
				defs := []jen.Code{jen.Id(firstValue.Name).Id(enum).Op("=").Iota()}
				for _, value := range values.Values[1:] {
					defs = append(defs, jen.Id(value.Name))
				}
				file.Const().Defs(defs...)
			}

			// Render file.
			divider := strings.Repeat("-", 80)
			fmt.Printf("Generating %s_service.go:\n", strings.ToLower(name))
			fmt.Println(divider)
			err = file.Render(os.Stdout)
			if err != nil {
				panic(err)
			}
			fmt.Println(divider + "\n\n\n\n")
		}
	}
}

func JenType(t Type, id *jen.Statement) jen.Code {
	switch t.Code {
	case "SINT32":
		return id.Int32()
	case "BOOL":
		return id.Bool()
	case "STRING":
		return id.String()
	case "FLOAT":
		return id.Float32()
	case "DOUBLE":
		return id.Float64()
	case "LIST":
		if len(t.Types) != 1 {
			panic(t)
		}
		return JenType(t.Types[0], id.Index())
	case "TUPLE":
		// TODO: generate tuple structs?
		// looks like we only ever have:
		//
		//   - (double, double): RectTransform_get_Position
		//   - (double, double, double): many
		//   - (double, double, double, double): Text_set_Rotation
		//   - ((double, double, double), (double, double, double)): Part_BoundingBox
		//
		// i think we can generate specific structs: twopoint, threepoint, fourpoint, boundingbox
		if len(t.Types) == 2 {
			if t.Types[0].Code == "TUPLE" {
				return id.Id("BoundingBox")
			}
			return id.Id("Point")
		} else if len(t.Types) == 3 {
			return id.Id("Vector")
		} else if len(t.Types) == 4 {
			return id.Id("Quaternion")
		} else {
			panic(t)
		}
	case "ENUMERATION":
		return id.Id(t.Name)
	case "CLASS":
		return id.Id("*" + t.Name)
	default:
		panic(t.Code)
	}
	return nil
}
