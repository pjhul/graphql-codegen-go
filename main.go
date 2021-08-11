package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

 func FormatSelectionSet(selectionSet ast.SelectionSet, depth int) string {
		if len(selectionSet) == 0 {
				return ""
		}

		var sb strings.Builder

		for _, selection := range selectionSet {
				switch selection := selection.(type) {
				case *ast.Field:
						for i := 0; i <= depth; i++ {
								sb.WriteString("    ")
						}
						if len(selection.SelectionSet) == 0 {
								sb.WriteString(
										strings.Title(selection.Name) + " " + FormatType(selection.Definition.Type) + " \x60json:\"" + selection.Name + "\"\x60\n",
								)
						} else {
								sb.WriteString(
										strings.Title(selection.Name) + " struct {\n" + FormatSelectionSet(selection.SelectionSet, depth + 1),
								)

								for i := 0; i <= depth; i++ {
										sb.WriteString("    ")
								}

								sb.WriteString(
										"} \x60json:\"" + selection.Name + "\"\x60\n",
								)
						}
				case *ast.FragmentSpread:
						if len(selection.Definition.SelectionSet) > 0 {
								sb.WriteString(
										FormatSelectionSet(selection.Definition.SelectionSet, depth),
								)
						}
				case *ast.InlineFragment:
						if len(selection.SelectionSet) > 0 {
								sb.WriteString(
										FormatSelectionSet(selection.SelectionSet, depth),
								)
						}
				default:
				}
		}

		return sb.String()
}

func FormatType(t *ast.Type) string {
		var sb strings.Builder

		if !t.NonNull {
				sb.WriteString("*")
		}

		if t.Elem != nil {
				sb.WriteString("[]" + t.Elem.NamedType)
		} else {
				sb.WriteString(t.NamedType)
		}

		return sb.String()
}

func main() {
		schemaFile, err := ioutil.ReadFile("./schema.graphql")
		if err != nil { panic(err) }

		source := ast.Source{
				Name: "schema.graphql",
				Input: string(schemaFile),
				BuiltIn: true,
		}

		schema, gqlErr := gqlparser.LoadSchema(&source)
		if gqlErr != nil {
				fmt.Println(gqlErr)
				os.Exit(1)
		}

		tmpl, err := template.New("schema.gotpl").Funcs(template.FuncMap{
				"goName": func(name string) string {
						return strings.Title(name)
				},
				"goScalar": func(name string) string {
						typeMap := map[string]string{
								"ID":						"uint64",
								"Float":				"float64",
								"Boolean":			"bool",
								"String":				"string",
								"Int":					"uint64",
								"_text":				"[]string",
								"timestamptz":	"string",
						}

						newType, ok := typeMap[name]
						if ok {
								return newType
						} else {
								return name
						}
				},
				"goType": FormatType,
		}).ParseFiles("schema.gotpl")

		f, err := os.Create("schema.go")
		if err != nil { panic(err) }
		defer f.Close()

		if err != nil { panic(err) }
		err = tmpl.Execute(f, schema)
		if err != nil { panic(err) }

		document := `
				fragment vendor on Vendor {
						id
						name
						createdAt
				}

				fragment test on Product {
						... on Product {
								id
								vendor {
										...vendor
										userId
								}
								createdAt
						}
						updatedAt
				}

				query GetProduct($id: Int!) {
						productById(id: $id) {
								...test
						}
				}

		`

		query := gqlparser.MustLoadQuery(schema, document)

		// fmt.Println(query.Operations[0].SelectionSet[0].(*ast.Field).SelectionSet[0].(*ast.FragmentSpread).ObjectDefinition.Name)

		tmpl, err = template.New("operations.gotpl").Funcs(template.FuncMap{
				"goName": func(name string) string {
						return strings.Title(name)
				},
				"goScalar": func(name string) string {
						typeMap := map[string]string{
								"ID":						"uint64",
								"Float":				"float64",
								"Boolean":			"bool",
								"String":				"string",
								"Int":					"uint64",
								"_text":				"[]string",
								"timestamptz":	"time.Time",
						}

						newType, ok := typeMap[name]
						if ok {
								return newType
						} else {
								return name
						}
				},
				"goType": func(t *ast.Type) string {
						var sb strings.Builder

						if !t.NonNull {
								sb.WriteString("*")
						}

						if t.Elem != nil {
								sb.WriteString("[]" + t.Elem.NamedType)
						} else {
								sb.WriteString(t.NamedType)
						}

						return sb.String()
				},
				"formatFragmentName": func(name string) string {
						return strings.Title(name) + "Fragment"
				},
				"formatSelectionSet": FormatSelectionSet,
				"rawQuery": func(op *ast.OperationDefinition, fragments ast.FragmentDefinitionList) string {
						var sb strings.Builder

						fmt.Println(op, fragments)

						f := Formatter{
								Writer: &sb,
						}

						f.FormatOperationDefinition(op)
						f.FormatFragmentDefinitionList(fragments)

						return sb.String()
				},
		}).ParseFiles("operations.gotpl")

		f, err = os.Create("operations.go")
		if err != nil { panic(err) }
		defer f.Close()

		if err != nil { panic(err) }
		err = tmpl.Execute(f, query)
		if err != nil { panic(err) }*/
}

