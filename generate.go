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

//go:embed templates/inputs.gotpl
var inputsTmpl string

//go:embed templates/operations.gotpl
var operationsTmpl string

func generateInputs(schema *ast.Schema, out io.Writer) error {
		fmt.Println("Generating input types...")

		tmpl, err := template.New("inputs.gotpl").Funcs(template.FuncMap{
				"formatName": formatName,
				"formatScalar": formatScalar,
				"formatType": formatType,
		}).Parse(inputsTmpl)

		err = tmpl.Execute(out, schema)
		if err != nil {
				return err
		}

		return nil
}

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

func generateOperations(schema *ast.Schema, queryDoc *ast.QueryDocument, out io.Writer) error {
		fmt.Println("Generating operations...")

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

func parseQueryDocuments(schema *ast.Schema, documents []string) (*ast.QueryDocument, error) {
		var parentDoc ast.QueryDocument
		parentDoc.Operations = ast.OperationList{}
		parentDoc.Fragments = ast.FragmentDefinitionList{}

		for _, document := range documents {
				queryDoc, err := gqlparser.LoadQuery(schema, document)
				if err != nil {
						return nil, err
				}

				operations := []*ast.OperationDefinition{}
				for _, op := range queryDoc.Operations {
						operations = append(operations, inlineOperationDefinition(queryDoc, op))
				}

				parentDoc.Operations = append(parentDoc.Operations, operations...)
				parentDoc.Fragments = append(parentDoc.Fragments, queryDoc.Fragments...)
		}

		return &parentDoc, nil
}

func inlineOperationDefinition(queryDoc *ast.QueryDocument, operation *ast.OperationDefinition) *ast.OperationDefinition {
		operation.SelectionSet = *inlineSelectionSet(queryDoc, &operation.SelectionSet)
		return operation
}

func inlineSelectionSet(queryDoc *ast.QueryDocument, selectionSet *ast.SelectionSet) *ast.SelectionSet {
		if selectionSet == nil || len(*selectionSet) == 0 {
				return nil
		}

		inlined := ast.SelectionSet{}
		for _, selection := range *selectionSet {
				switch selection := selection.(type) {
				case *ast.Field:
						if len(selection.SelectionSet) > 0 {
								selection.SelectionSet = *inlineSelectionSet(queryDoc, &selection.SelectionSet)
								inlined = append(inlined, selection)
						} else {
								inlined = append(inlined, selection)
						}
				case *ast.FragmentSpread:
						inlined = append(inlined, *inlineSelectionSet(queryDoc, &selection.Definition.SelectionSet)...)
				case *ast.InlineFragment:
						inlined = append(inlined, *inlineSelectionSet(queryDoc, &selection.SelectionSet)...)
				default:

				}
		}

		return &inlined
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

		return sb.String()
}

func formatSelectionSet(selectionSet ast.SelectionSet, depth int) string {
		if len(selectionSet) == 0 { return "" }

		var sb strings.Builder

		// this recursion feels overcomplicated
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

		return sb.String()
}
