package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/ilikebits/jeb/cmd/krpc-gen/service"
)

const (
	protoImportPath    = "github.com/golang/protobuf/proto"
	wrappersImportPath = "github.com/golang/protobuf/ptypes/wrappers"
	pbImportPath       = "github.com/ilikebits/jeb/krpc/pb"
)

func main() {
	// Parse flags.
	dir := flag.String("dir", "", "directory of JSON service definitions (required)")
	out := flag.String("out", "", "import path of generated package")
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
		var services service.Services
		err = json.Unmarshal(contents, &services)
		if err != nil {
			panic(err)
		}

		// Generate service client.
		for serviceName, definition := range services {
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
			statics := make(map[string]map[string]service.Procedure)
			methods := make(map[string]map[string]service.Procedure)
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
				var table map[string]service.Procedure
				if splits[1] == "static" {
					table = statics[class]
					if table == nil {
						statics[class] = make(map[string]service.Procedure)
						table = statics[class]
					}
				} else {
					table = methods[class]
					if table == nil {
						methods[class] = make(map[string]service.Procedure)
						table = methods[class]
					}
				}
				// Set procedure.
				table[name] = proc
			}

			// New file.
			file := NewFilePath("github.com/ilikebits/jeb/krpc")

			// Service singleton.
			file.Type().Id(serviceName).Struct(
				Id("conn").Op("*").Id("Conn"),
			)

			// Classes.
			for class := range definition.Classes {
				// Define class struct.
				file.Type().Id(class).Struct(
					Id("conn").Op("*").Id("Conn"), // Should this be a reference back to the service instead?
					Id("id").Uint64(),
				)

				// Define static class struct and methods.
				if len(statics[class]) > 0 {
					file.Type().Id(class + "Static").Struct(
						Id("conn").Op("*").Id("Conn"),
					)
					for name, proc := range statics[class] {
						// Take idiomatic parameters.
						var params []Code
						var args []Code
						var marshallers [][]Code
						for i, param := range proc.Parameters {
							info := GenerateType(param.Type)
							marshalled, marshal := info.Marshal(param.Name)

							params = append(params, Id(param.Name).Add(info.Type))
							args = append(args, Op("&").Qual(pbImportPath, "Argument").Values(Dict{
								Id("Position"): Lit(i),
								Id("Value"):    marshalled,
							}))
							marshallers = append(marshallers, marshal)
						}

						static := file.Func().Params(
							Id("static").Op("*").Id(class + "Static"),
						).Id(strings.TrimPrefix(name, class+"_static_")).Params(params...)

						// Generate return type.
						var returns []Code
						var zeros []Code
						var unmarshaller TypeGenerator
						if proc.ReturnType.Code != "" {
							info := GenerateType(proc.ReturnType)
							returns = append(returns, info.Type)
							zeros = append(zeros, info.Zero)
							unmarshaller = info.Unmarshal
						}
						returns = append(returns, Error())
						static.Params(returns...)

						errReturn := If(Err().Op("!=").Nil().Block(Return(append(zeros, Err())...)))

						// Generate method body.
						var block []Code
						// Marshal arguments.
						for _, marshal := range marshallers {
							if len(marshal) > 0 {
								block = append(block, marshal...)
								block = append(block, errReturn)
							}
						}
						// Construct request.
						block = append(block, Id("req").Op(":=").Qual(pbImportPath, "Request").Values(Dict{
							Id("Calls"): Index().Op("*").Qual(pbImportPath, "ProcedureCall").Values(
								Op("&").Qual(pbImportPath, "ProcedureCall").Values(Dict{
									Id("Service"):           Lit(serviceName),
									Id("service.Procedure"): Lit(name),
									Id("Arguments"):         Index().Op("*").Qual(pbImportPath, "Argument").Values(List(args...)),
								}),
							),
						}))
						// Make request.
						block = append(block,
							List(Id("_"), Err()).Op("=").
								Id("static").Dot("conn").Dot("Send").Call(Op("&").Id("req")))
						block = append(block, errReturn)
						// Read response.
						block = append(block, Id("res").Op(":=").Qual(pbImportPath, "Response").Values())
						block = append(block, Err().Op("=").Id("static").Dot("conn").Dot("Read").Call(Op("&").Id("res")))
						block = append(block, errReturn)
						block = append(block, If(Id("e").Op(":=").Id("res").Dot("GetError").Call(), Id("e").Op("!=").Nil()).Block(
							Return(append(zeros, Qual("errors", "New").Call(Id("e").Dot("GetDescription").Call()))...),
						))
						// Unmarshal result.
						if len(returns) == 1 {
							continue
						}
						// Return result errors.
						block = append(block, Id("result").Op(":=").Id("res").Dot("GetResults").Call().Index(Lit(0)))
						block = append(block, If(Id("e").Op(":=").Id("result").Dot("GetError").Call(), Id("e").Op("!=").Nil()).Block(
							Return(append(zeros, Qual("errors", "New").Call(Id("e").Dot("GetDescription").Call()))...),
						))
						// Unmarshal result bytes.
						block = append(block, Id("resultBytes").Op(":=").Id("result").Dot("GetValue").Call())
						result, steps := unmarshaller("resultBytes")
						block = append(block, steps...)
						if len(steps) > 0 {
							block = append(block, errReturn)
						}
						if proc.ReturnType.Code == "CLASS" {
							block = append(block, Id("resultStruct").Op(":=").Add(result))
							block = append(block, Id("resultStruct").Dot("conn").Op("=").Id("static").Dot("conn"))
							block = append(block, Return(Op("&").Id("resultStruct"), Nil()))
						} else {
							block = append(block, Return(result, Nil()))
						}
						static.Block(block...)
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
				defs := []Code{Id(firstValue.Name).Id(enum).Op("=").Iota()}
				for _, value := range values.Values[1:] {
					defs = append(defs, Id(enum+value.Name))
				}
				file.Const().Defs(defs...)
			}

			// Render file.
			if *out == "" {
				divider := strings.Repeat("-", 80)
				fmt.Printf("Generating %s_service.go:\n", strings.ToLower(serviceName))
				fmt.Println(divider)

				var rendered bytes.Buffer
				err = file.Render(&rendered)
				var annotated string
				for i, line := range strings.Split(rendered.String(), "\n") {
					annotated += fmt.Sprintf("%4d | %s\n", i+1, strings.Replace(line, "\t", "  ", -1))
				}
				fmt.Println(annotated)

				if err != nil {
					// Special case: make debugging formatting errors easier.
					msg := err.Error()
					if strings.Contains(msg, "while formatting source:") {
						splits := strings.Split(msg, "while formatting source:")

						for i, line := range strings.Split(splits[1], "\n") {
							annotated += fmt.Sprintf("%4d | %s\n", i+1, strings.Replace(line, "\t", "  ", -1))
						}
						log.Println(annotated)
						log.Println(splits[0] + "while formatting source")
					} else {
						panic(err.Error())
					}
				}
				fmt.Println(divider + "\n\n\n\n")
			} else {
				file.Save(filepath.Join(os.Getenv("GOPATH"), "src", *out, "generated_"+strings.ToLower(serviceName)+"_service.go"))
			}
		}
	}
}

type TypeGenerator func(name string) (result Code, steps []Code)

type TypeInfo struct {
	Type, Zero Code

	Marshal   TypeGenerator
	Unmarshal TypeGenerator
}

func GenerateType(t service.Type) TypeInfo {
	// Generate parameter type, zero, and marshalling code.
	switch t.Code {
	case "SINT32":
		return TypeInfo{
			Type:      Int32(),
			Zero:      Lit(0),
			Marshal:   MarshalBuffer("EncodeZigzag32"),
			Unmarshal: UnmarshalBuffer("DecodeZigzag32"),
		}
	case "BOOL":
		return TypeInfo{
			Type: Bool(),
			Zero: Lit(false),
			Marshal: func(paramName string) (Code, []Code) {
				return Qual(protoImportPath, "EncodeVarint").Call(Id(paramName)), nil
			},
			Unmarshal: func(byteSlice string) (Code, []Code) {
				decoded := byteSlice + "Decoded"
				steps := []Code{List(Id(decoded), Err()).Op(":=").Qual(protoImportPath, "DecodeVarint").Call(Id(byteSlice))}
				return Id(decoded), steps
			},
		}
	case "STRING":
		return TypeInfo{
			Type:      String(),
			Zero:      Lit(""),
			Marshal:   MarshalBuffer("EncodeStringBytes"),
			Unmarshal: UnmarshalBuffer("DecodeStringBytes"),
		}
	case "FLOAT":
		return TypeInfo{
			Type:      Float32(),
			Zero:      Lit(0.0),
			Marshal:   MarshalBuffer("EncodeFixed32"),
			Unmarshal: UnmarshalBuffer("DecodeFixed32"),
		}
	case "DOUBLE":
		return TypeInfo{
			Type:      Float64(),
			Zero:      Lit(0.0),
			Marshal:   MarshalBuffer("EncodeFixed64"),
			Unmarshal: UnmarshalBuffer("DecodeFixed64"),
		}
	case "LIST":
		if len(t.Types) != 1 {
			panic(t)
		}
		return TypeInfo{
			Type: Id("TODO"),
			Zero: Id("TODO").Values(),
			Marshal: func(_ string) (Code, []Code) {
				return Id("TODO"), nil
			},
			Unmarshal: func(byteSlice string) (Code, []Code) {
				return Id("TODO"), nil
			},
		}
	case "TUPLE":
		// Since Go has neither generics nor first-class support for tuples, we
		// must generate tuple structs for each type of tuple. As you can imagine,
		// this is very annoying. For pragmatism, we hard-code a set of 4 tuple
		// structs that we use in generation, since it appears these are the only
		// 4 types of tuples that ever occur:
		//
		//   - (double, double): RectTransform_get_Position
		//   - (double, double, double): many
		//   - (double, double, double, double): Text_set_Rotation
		//   - ((double, double, double), (double, double, double)): Part_BoundingBox
		var tupleStruct string

		if len(t.Types) == 2 {
			if t.Types[0].Code == "TUPLE" {
				tupleStruct = "BoundingBox"
			}
			tupleStruct = "Point"
		} else if len(t.Types) == 3 {
			tupleStruct = "Vector"
		} else if len(t.Types) == 4 {
			tupleStruct = "Quaternion"
		} else {
			panic(t)
		}

		return TypeInfo{
			Type: Id(tupleStruct),
			Zero: Id(tupleStruct).Values(),
			Marshal: func(_ string) (Code, []Code) {
				return Id("TODO"), nil
			},
			Unmarshal: func(byteSlice string) (Code, []Code) {
				return Id("TODO"), nil
			},
		}
	case "ENUMERATION":
		return TypeInfo{
			Type: Id(t.Name),
			Zero: Lit(-1),
			Marshal: func(paramName string) (Code, []Code) {
				return Qual(protoImportPath, "EncodeVarint").Call(Id(paramName)), nil
			},
			Unmarshal: func(byteSlice string) (Code, []Code) {
				decoded := byteSlice + "Decoded"
				steps := []Code{List(Id(decoded), Err()).Op(":=").Qual(protoImportPath, "DecodeVarint").Call(Id(byteSlice))}
				return Id(t.Name).Call(Id(decoded)), steps
			},
		}
	case "CLASS":
		return TypeInfo{
			Type: Op("*").Id(t.Name),
			Zero: Nil(),
			Marshal: func(paramName string) (Code, []Code) {
				return Qual(protoImportPath, "EncodeVarint").Call(Id(paramName)), nil
			},
			Unmarshal: func(byteSlice string) (Code, []Code) {
				decoded := byteSlice + "Decoded"
				steps := []Code{List(Id(decoded), Err()).Op(":=").Qual(protoImportPath, "DecodeVarint").Call(Id(byteSlice))}
				return Id(t.Name).Values(Dict{Id("id"): Id(decoded)}), steps
			},
		}
	}
	panic(t.Code)
}

func MarshalBuffer(method string) TypeGenerator {
	return func(paramName string) (Code, []Code) {
		buf := paramName + "Buffer"
		steps := []Code{
			Id(buf).Op(":=").Qual(protoImportPath, "Buffer").Values(),
			Err().Op(":=").Add(Id(buf).Dot(method).Call(Id(paramName))),
		}
		result := Id(buf).Dot("Bytes").Call()
		return result, steps
	}
}

func UnmarshalBuffer(method string) TypeGenerator {
	return func(byteSlice string) (Code, []Code) {
		parsed := byteSlice + "Parsed"
		steps := []Code{
			List(Id(parsed), Err()).Op(":=").Qual(protoImportPath, "NewBuffer").Call(Id(byteSlice)).Dot(method).Call(),
		}
		result := Id(parsed)
		return result, steps
	}
}
