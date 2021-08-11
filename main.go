package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"modosuite/graphql-codegen-go/gofmt"

	"github.com/vektah/gqlparser/v2/ast"
)

var (
		schemaPath = flag.String("schema", "", "Path to locate the graphql schema")
		// operationsPath := flag.String("ops", "", "Path to locate the graphql operations")
)

func main() {
		flag.Parse()

		schemaFile, err := ioutil.ReadFile(*schemaPath)
		if err != nil { panic(err) }
		source := &ast.Source{
				Name: *schemaPath,
				Input: string(schemaFile),
				BuiltIn: false,
		}

		var buf bytes.Buffer

		err = generateSchema(source, &buf)
		if err != nil { panic(err) }

		file, err := os.Create("schema.go")
		if err != nil { panic(err) }

		defer file.Close()

		gofmt.ProcessFile("./schema.go", &buf, file, false)

		if err != nil { panic(err) }

		// w.WriteAt(bytes, 0)

		fmt.Println("Successfully generated scheme file!")
}
