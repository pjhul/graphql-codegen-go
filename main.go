package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"modosuite/graphql-codegen-go/gofmt"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
		schemaPath = flag.String("schema", "", "Path to locate the graphql schema")
		operationsGlob = flag.String("operations", "", "Glob to locate the graphql operations")
)

func main() {
		flag.Parse()

		var buf bytes.Buffer

		fmt.Printf("Reading schema from %s\n", *schemaPath)
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

		operationFiles, err := filepath.Glob(*operationsGlob)
		if err != nil { panic(err) }

		for _, file := range operationFiles {
				opFile, err := ioutil.ReadFile(file)
				if err != nil { panic(err) }

				err = generateOperations(schemaDoc, string(opFile), &buf)
				if err != nil { panic(err) }
		}

		file, err := os.Create("schema.go")
		if err != nil { panic(err) }
		defer file.Close()

		err = gofmt.ProcessFile("./schema.go", &buf, file, false)
		if err != nil { panic(err) }

		fmt.Println("Successfully generated schema file!")
}
