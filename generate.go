package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"

	_ "embed"
)

var typeMap = map[string]string{
		"ID":						"string",
		"Int":					"int64",
		"Float":				"float64",
		"Boolean":			"bool",
		"String":				"string",
		"_text":				"[]string",
		"timestamptz":	"string",
}

//go:embed templates/schema.gotpl
var schemaTmpl string

//go:embed templates/operations.gotpl
var operationsTmpl string

func generateSchema(schema *ast.Schema, out io.Writer) error {
		fmt.Println("Generating schema types...")

		tmpl, err := template.New("schema.gotpl").Funcs(template.FuncMap{
				"formatName": formatName,
				"formatScalar": formatScalar,
				"formatType": formatType,
		}).Parse(schemaTmpl)

		err = tmpl.Execute(out, schema)
		if err != nil {
				return err
		}

		return nil
}

func generateOperations(schema *ast.Schema, document string, out io.Writer) error {
		fmt.Println("Generating operations...")

		queryDoc, gqlErr := gqlparser.LoadQuery(schema, document)
		if gqlErr != nil {
				return gqlErr
		}

		tmpl, err := template.New("operations.gotpl").Funcs(template.FuncMap{
				"formatName": formatName,
				"formatScalar": formatScalar,
				"formatType": formatType,
				"formatFragmentName": formatFragmentName,
				"formatSelectionSet": formatSelectionSet,
				"formatQuery": formatQuery,
		}).Parse(operationsTmpl)

		err = tmpl.Execute(out, queryDoc)
		if err != nil {
				return err
		}

		return nil
}

func formatName(name string) string {
		return strings.Title(name)
}

func formatFragmentName(name string) string {
		return strings.Title(name) + "Fragment"
}

func formatScalar(scalar string) string {
		newType, ok := typeMap[scalar]

		if ok {
				return newType
		} else {
				return "string"
		}
}

func formatType(t *ast.Type) string {
		var sb strings.Builder

		if !t.NonNull {
				sb.WriteString("*")
		}

		if t.Elem != nil {
				newType, ok := typeMap[t.Elem.NamedType]

				if ok {
						sb.WriteString("[]" + newType)
				} else {
						sb.WriteString("[]" + t.Elem.NamedType)
				}
		} else {
				newType, ok := typeMap[t.NamedType]

				if ok {
						sb.WriteString(newType)
				} else {
						sb.WriteString(t.NamedType)
				}
		}

		return t.String()
}

func formatSelectionSet(selectionSet ast.SelectionSet, depth int) string {
		if len(selectionSet) == 0 { return "" }

		var sb strings.Builder

		for _, selection := range selectionSet {
				switch selection := selection.(type) {
				case *ast.Field:
						for i := 0; i <= depth; i++ {
								sb.WriteString("    ")
						}

						// FIXME: Does not take aliases into account

						if len(selection.SelectionSet) == 0 {
								sb.WriteString(
										strings.Title(selection.Name) + " " + formatType(selection.Definition.Type) + " `json:\"" + selection.Name + "\"`\n",
								)
						} else {
								sb.WriteString(strings.Title(selection.Name) + " ")

								if !selection.Definition.Type.NonNull {
										sb.WriteString("*")
								}

								if selection.Definition.Type.Elem != nil {
										sb.WriteString("[]")
								}

								sb.WriteString(
										"struct {\n" + formatSelectionSet(selection.SelectionSet, depth + 1),
								)

								for i := 0; i <= depth; i++ {
										sb.WriteString("    ")
								}

								sb.WriteString(
										"} `json:\"" + selection.Name + "\"`\n",
								)
						}
				case *ast.FragmentSpread:
						if len(selection.Definition.SelectionSet) > 0 {
								sb.WriteString(
										formatSelectionSet(selection.Definition.SelectionSet, depth),
								)
						}
				case *ast.InlineFragment:
						if len(selection.SelectionSet) > 0 {
								sb.WriteString(
										formatSelectionSet(selection.SelectionSet, depth),
								)
						}
				default:
				}
		}

		return sb.String()
}

func formatQuery(op *ast.OperationDefinition, fragments ast.FragmentDefinitionList) string {
		var sb strings.Builder

		f := Formatter{Writer: &sb}

		f.FormatOperationDefinition(op)
		f.FormatFragmentDefinitionList(fragments)

		return sb.String()
}
