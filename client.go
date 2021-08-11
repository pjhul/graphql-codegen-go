package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// TODO: Create Stripe style sdk
/*type API struct {
		Platforms *platform.Client
}*/

type AdminClient struct {
		Endpoint		string
		AdminSecret string
}

type GraphQLVariables map[string]interface{}

type GraphQLErrorExtensions struct {
		Code string `json:"code"`
		Path string `json:"path"`
}

type GraphQLError struct {
		Message			string									`json:"message"`
		Extensions	GraphQLErrorExtensions	`json:"extensions"`
}

type GraphQLErrors []GraphQLError

func (errs GraphQLErrors) Error() string {
		var sb strings.Builder
		sb.WriteString("ERROR\n")

		for _, err := range errs {
				sb.WriteString(
						fmt.Sprintf("%s ON %s -> %s", err.Extensions.Code, err.Extensions.Path, err.Message),
				)
		}

		return sb.String()
}

type GraphQLResult struct {
		Data		json.RawMessage			`json:"data"`
		Errors	GraphQLErrors				`json:"errors"`
}

func (c *AdminClient) Request(
		query string,
		variables map[string]interface{},
) (*GraphQLResult, error) {
		body, err := json.Marshal(map[string]interface{}{
				"query": query,
				"variables": variables,
		})
		if err != nil {
				return nil, err
		}

		request, err := http.NewRequest(
				"POST",
				c.Endpoint,
				bytes.NewBuffer(body),
		)

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Hasura-Admin-Secret", c.AdminSecret)

		// bytes, err := httputil.DumpRequest(request, true)
		// fmt.Printf("%s\n", string(bytes))

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
				return nil, err
		}
		defer response.Body.Close()

		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
				return nil, err
		}

		var result GraphQLResult
		err = json.Unmarshal(body, &result)
		if err != nil {
				return nil, err
		}

		if result.Errors == nil {
				return &result, nil
		} else if len(result.Errors) > 0 {
				return nil, result.Errors
		}

		return nil, errors.New("Unknown error occured")
}
