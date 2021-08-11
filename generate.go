package main

import (
	"html/template"
	"io"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var typeMap = map[string]string{
		"ID":						"uint64",
		"Float":				"float64",
		"Boolean":			"bool",
		"String":				"string",
		"Int":					"uint64",
		"_text":				"[]string",
		"timestamptz":	"string",
}

func generateSchema(source *ast.Source, w io.Writer) error {
		schema, gqlErr := gqlparser.LoadSchema(source)
		if gqlErr != nil {
				return gqlErr
		}

		tmpl, err := template.New("schema.gotpl").Funcs(template.FuncMap{
				"formatName": formatName,
				"formatScalar": formatScalar,
				"formatType": formatType,
		}).ParseFiles("./templates/schema.gotpl")

		err = tmpl.Execute(w, schema)
		if err != nil {
				return err
		}

		/*query := gqlparser.MustLoadQuery(schema, document)

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
		return nil
}

func formatName(name string) string {
		return strings.Title(name)
}

func formatScalar(scalar string) string {
		newType, ok := typeMap[scalar]

		if ok {
				return newType
		} else {
				return scalar
		}
}

func formatType(t *ast.Type) string {
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

func formatSelectionSet(selectionSet ast.SelectionSet, depth int) string {
		if len(selectionSet) == 0 { return "" }

		var sb strings.Builder

		for _, selection := range selectionSet {
				switch selection := selection.(type) {
				case *ast.Field:
						for i := 0; i <= depth; i++ {
								sb.WriteString("    ")
						}
						if len(selection.SelectionSet) == 0 {
								sb.WriteString(
										strings.Title(selection.Name) + " " + formatType(selection.Definition.Type) + " \x60json:\"" + selection.Name + "\"\x60\n",
								)
						} else {
								sb.WriteString(
										strings.Title(selection.Name) + " struct {\n" + formatSelectionSet(selection.SelectionSet, depth + 1),
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
