package main

import (
	"bytes"
	"flag"
	"fmt"
	"modosuite/graphql-codegen-go/gofmt"
	"os"
)

var (
		packageName = flag.String("package", "main", "Name of the package to output")
		schemaPath = flag.String("schema", "", "Path to locate the graphql schema")
		operationsGlob = flag.String("operations", "", "Glob to locate the graphql operations")
		url	= flag.String("url", "", "Endpoint of the api")
)

func main() {
		flag.Parse()

		var buf bytes.Buffer

		buf.WriteString(fmt.Sprintf("package %s\n", *packageName))

		schema, err := Introspect("https://api.dev.modosuite.com/v1/graphql")
		if err != nil {
				fmt.Println(err)
		}

		err = generateSchema(schema, &buf)
		if err != nil { panic(err) }

		file, err := os.Create("schema.go")
		if err != nil { panic(err) }
		defer file.Close()

		err = gofmt.ProcessFile("./schema.go", &buf, file, false)
		if err != nil { panic(err) }

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

		fmt.Println("Successfully generated schema file!")*/
}
