//go:build ignore
// +build ignore

// This code generates the elephp.schema.json and default_config.json files.

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/danielgtaylor/huma/schema"
	"github.com/laytan/elephp/internal/config"
)

const (
	SchemaVersion = "https://json-schema.org/draft/2020-12/schema"
	Title         = "ElePHP"
	Description   = "Configuration format for the ElePHP language server."
)

type Schema struct {
	SchemaVersion string `json:"$schema"`
	Id            string `json:"$id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	*schema.Schema
}

func main() {
	dumpSchema()
	dumpDefaultConfig()
}

func dumpSchema() {
	s, err := schema.Generate(reflect.TypeOf(&config.Schema{}))
	if err != nil {
		panic(err)
	}

	fh, err := os.Create("elephp.schema.json")
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	fullSchema := &Schema{
		SchemaVersion: SchemaVersion,
		Id:            config.SchemaLocation,
		Title:         Title,
		Description:   Description,
		Schema:        s,
	}

	enc := json.NewEncoder(fh)
	enc.SetIndent("", "    ")
	if err = enc.Encode(fullSchema); err != nil {
		panic(err)
	}
}

func dumpDefaultConfig() {
	cfg := config.DefaultWithoutComputed()
	fh, err := os.Create("default_config.json")
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	// We want to also show the "empty" fields in the default config.
	dcfg := removeOmitEmpty(*cfg)

	enc := json.NewEncoder(fh)
	enc.SetIndent("", "    ")
	if err = enc.Encode(dcfg); err != nil {
		panic(err)
	}
}

// removeOmitEmpty naively turns the given type into a map[string]any,
// ignoring any 'omitempty' tags, but still using the name from the json tag.
// obj needs to be a pointer to a struct.
func removeOmitEmpty(obj any) map[string]any {
	typ, val := reflect.TypeOf(obj), reflect.ValueOf(obj)
	out := make(map[string]any, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		key := field.Name

		// See if there is json tag to take the name of.
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			if jsonName, _, ok := strings.Cut(jsonTag, ","); ok && jsonName != "" {
				key = jsonName
			} else if jsonTag == "-" {
				continue
			}
		}

		if key == "-" {
			continue
		}

		// Recursively do this.
		if field.Type.Kind() == reflect.Struct {
			value := removeOmitEmpty(val.Field(i).Interface())
			// Handle embedded fields.
			if field.Anonymous {
				for k, v := range value {
					out[k] = v
				}
			} else {
				out[key] = value
			}
			continue
		}

		value := val.Field(i).Interface()
		out[key] = value
	}

	return out
}
