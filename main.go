package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"modosuite/graphql-codegen-go/gofmt"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
		schemaPath = flag.String("schema", "", "Path to locate the graphql schema")
		operationPath = flag.String("ops", "", "Path to locate the graphql operations")
)

func main() {
		flag.Parse()

		schemaFile, err := ioutil.ReadFile(*schemaPath)
		if err != nil { panic(err) }

		opFile, err := ioutil.ReadFile(*operationPath)
		if err != nil { panic(err) }

		source := &ast.Source{
				Name: *schemaPath,
				Input: string(schemaFile),
				BuiltIn: false,
		}

		schemaDoc, gqlErr := gqlparser.LoadSchema(source)
		if gqlErr != nil { panic(gqlErr) }

		var buf bytes.Buffer

		err = generateSchema(schemaDoc, &buf)
		if err != nil { panic(err) }

		err = generateOperations(schemaDoc, string(opFile), &buf)
		if err != nil { panic(err) }

		file, err := os.Create("schema.go")
		if err != nil { panic(err) }
		defer file.Close()

		err = gofmt.ProcessFile("./schema.go", &buf, file, false)
		if err != nil { panic(err) }

		fmt.Println("Successfully generated schema file!")
}
