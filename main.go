package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"modosuite/graphql-codegen-go/gofmt"
	"os"
	"path/filepath"
	"strings"
)

type headers []string

func (i *headers) String() string {
    return "my string representation"
}

func (i *headers) Set(value string) error {
    *i = append(*i, value)
    return nil
}

var (
		packageName = flag.String("package", "main", "Name of the package to output")
		schemaPath = flag.String("schema", "", "Path to locate the graphql schema")
		operationsGlob = flag.String("operations", "", "Glob to locate the graphql operations")
		endpoint	= flag.String("E", "", "Endpoint of the api")
		fullSchema = flag.Bool("full", false, "Include full schema types")
)

var headerList headers

func main() {
		flag.Var(&headerList, "H", "")
		flag.Parse()

		var buf bytes.Buffer

		buf.WriteString(fmt.Sprintf(`
				package %s

				import (
						"encoding/json"
				)
		`, *packageName))

		headerMap := make(map[string][]string)
		for _, h := range headerList {
				slices := strings.Split(h, ":")
				headerMap[strings.TrimSpace(slices[0])] = []string{strings.TrimSpace(slices[1])}
		}

		schema, err := Introspect(*endpoint, headerMap)
		if err != nil {
				fmt.Println(err)
		}

		err = generateInputs(schema, &buf)
		if err != nil { panic(err) }

		if *fullSchema {
				err = generateSchema(schema, &buf)
				if err != nil { panic(err) }
		}

		operationFiles, err := filepath.Glob(*operationsGlob)
		if err != nil { panic(err) }

		queryDocs := []string{}
		for _, file := range operationFiles {
				opFile, err := ioutil.ReadFile(file)
				if err != nil { panic(err) }

				queryDocs = append(queryDocs, string(opFile))
		}

		queryDoc, err := parseQueryDocuments(schema, queryDocs)

		err = generateOperations(schema, queryDoc, &buf)
		if err != nil { panic(err) }

		file, err := os.Create("schema.go")
		if err != nil { panic(err) }
		defer file.Close()

		err = gofmt.ProcessFile("schema.go", &buf, file, false)
		if err != nil { panic(err) }

		fmt.Println("Successfully generated schema file!")

		/*fmt.Printf("Reading schema from %s\n", *schemaPath)
		content, err := ioutil.ReadFile(*schemaPath)
		if err != nil { panic(err) }

		source := &ast.Source{
				Name: *schemaPath,
				Input: string(content),
				BuiltIn: false,
		}

		schemaDoc, gqlErr := gqlparser.LoadSchema(source)
		if gqlErr != nil { panic(gqlErr) }

		err = generateSchema(schemaDoc, &buf)
		if err != nil { panic(err) }

		file, err := os.Create("schema.go")
		if err != nil { panic(err) }
		defer file.Close()

		err = gofmt.ProcessFile("./schema.go", &buf, file, false)
		if err != nil { panic(err) }

		*/
}
