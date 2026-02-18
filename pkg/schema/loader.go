package schema

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func Validate(schemaPath string, doc any) ([]string, error) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	docLoader := gojsonschema.NewGoLoader(doc)
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return nil, fmt.Errorf("validate %s: %w", schemaPath, err)
	}
	if result.Valid() {
		return nil, nil
	}

	errs := make([]string, 0, len(result.Errors()))
	for _, e := range result.Errors() {
		errs = append(errs, e.String())
	}
	return errs, nil
}
