package main

import (
	"bytes"
	"encoding/json"

	"io/ioutil"
	"net/http"

	"github.com/vektah/gqlparser/v2/ast"
)

type DefinitionKind string
const (
		DEF_SCALAR				DefinitionKind = "SCALAR"
		DEF_OBJECT				DefinitionKind = "OBJECT"
		DEF_INTERFACE			DefinitionKind = "INTERFACE"
		DEF_UNION					DefinitionKind = "UNION"
		DEF_ENUM					DefinitionKind = "ENUM"
		DEF_INPUT_OBJECT	DefinitionKind = "INPUT_OBJECT"
)

type TypeKind string
const (
		TYPE_SCALAR					TypeKind = "SCALAR"
		TYPE_OBJECT					TypeKind = "OBJECT"
		TYPE_INTERFACE			TypeKind = "INTERFACE"
		TYPE_UNION					TypeKind = "UNION"
		TYPE_ENUM						TypeKind = "ENUM"
		TYPE_INPUT_OBJECT		TypeKind = "INPUT_OBJECT"
		TYPE_LIST						TypeKind = "LIST"
		TYPE_NON_NULL				TypeKind = "NON_NULL"
)

type IntrospectionQueryResult struct {
		Data struct {
				Schema struct {
						QueryType struct {
								Name *string `json:"name"`
						} `json:"queryType"`
						MutationType struct {
								Name *string `json:"name"`
						} `json:"mutationType"`
						SubscriptionType struct {
								Name *string `json:"name"`
						} `json:"subscriptionType"`
						Types []FullType `json:"types"`
				} `json:"__schema"`
		} `json:"data"`
}

type FullType struct {
		Kind DefinitionKind `json:"kind"`
		Name string `json:"name"`
		Description string `json:"description"`
		Fields []Field `json:"fields"`
}

type Field struct {
		Name string					`json:"name"`
		Description string	`json:"description"`
		Type *TypeRef				`json:"type"`
}

type TypeRef struct {
		Kind		TypeKind		`json:"kind"`
		Name		string			`json:"name"`
		OfType	*TypeRef		`json:"ofType"`
}

func parseType(typeRef *TypeRef) *ast.Type {
		if typeRef == nil {
				return nil
		}

		switch(TypeKind(typeRef.Kind)) {
				case TYPE_NON_NULL:
						if(typeRef.OfType == nil) {
								return ast.NonNullNamedType(typeRef.Name, nil)
						} else {
								ofType := parseType(typeRef.OfType)
								ofType.NonNull = true

								return ofType
						}
				case TYPE_OBJECT, TYPE_INPUT_OBJECT, TYPE_INTERFACE, TYPE_UNION, TYPE_ENUM:
						if(typeRef.OfType == nil) {
								return ast.NamedType(typeRef.Name, nil)
						} else {
								return parseType(typeRef.OfType)
						}
				case TYPE_LIST:
						ofType := parseType(typeRef.OfType)

						return &ast.Type{
								Elem: ofType,
						}
				case TYPE_SCALAR:
						return ast.NonNullNamedType(typeRef.Name, nil)
		}

		return nil
}

func Introspect(endpoint string) (*ast.Schema, error) {
		query := `
				query IntrospectionQuery {
						__schema {
								queryType { name }
								mutationType { name }
								subscriptionType { name }
								types {
										...FullType
								}
						}
				}

				fragment FullType on __Type {
						kind
						name
						description
						fields(includeDeprecated: true) {
								name
								description
								type {
										...TypeRef
								}
						}
				}

				fragment TypeRef on __Type {
						kind
						name
						ofType {
								kind
								name
								ofType {
										kind
										name
										ofType {
												kind
												name
												ofType {
														kind
														name
														ofType {
																kind
																name
																ofType {
																		kind
																		name
																}
														}
												}
										}
								}
						}
				}
		`

		body, err := json.Marshal(map[string]interface{}{
				"query": query,
		})
		if err != nil {
				return nil, err
		}

		request, err := http.NewRequest(
				"POST",
				endpoint,
				bytes.NewBuffer(body),
		)
		if err != nil {
				return nil, err
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Hasura-Admin-Secret", "Bt9oEbGLX1UbyCkS43cIrrA5GtDzgkaPVJ9D0Nq4")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
				return nil, err
		}
		defer response.Body.Close()

		body, _ = ioutil.ReadAll(response.Body)

		// fmt.Println(string(body))

		var result IntrospectionQueryResult
		err = json.Unmarshal(body, &result)
		if err != nil {
				return nil, err
		}

		resultSchema := result.Data.Schema
		schema := &ast.Schema{}

		// fmt.Println(resultSchema.Types)

		if resultSchema.QueryType.Name != nil {
				schema.Query = &ast.Definition{
						Name: *resultSchema.QueryType.Name,
				}
		}

		if resultSchema.MutationType.Name != nil {
				schema.Mutation = &ast.Definition{
						Name: *resultSchema.MutationType.Name,
				}
		}

		if resultSchema.SubscriptionType.Name != nil {
				schema.Subscription = &ast.Definition{
						Name: *resultSchema.SubscriptionType.Name,
				}
		}

		schema.Types = make(map[string]*ast.Definition)
		for _, fullType := range resultSchema.Types {
				fields := []*ast.FieldDefinition{}

				switch(DefinitionKind(fullType.Kind)) {
						case DEF_OBJECT, DEF_INTERFACE:
								for _, field := range fullType.Fields {
										fieldDefinition := &ast.FieldDefinition{
												Name: field.Name,
												Description: field.Description,
										}

										fieldDefinition.Type = parseType(field.Type)
										fields = append(fields, fieldDefinition)
								}
								break
				}

				schema.Types[fullType.Name] = &ast.Definition{
						Kind: ast.DefinitionKind(fullType.Kind),
						Name: fullType.Name,
						Description: fullType.Description,
						Fields: ast.FieldList(fields),
				}
		}

		// fmt.Printf("%+v\n", schema)
		/*fmt.Printf("%+v\n", schema.Types["Vendor_aggregate_fields"].Fields)*/

		return schema, nil
}
