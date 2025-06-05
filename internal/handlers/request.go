package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// validator is an object that can be validated.
type validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

// decodeValid decodes a model from a http request and performs validation
// on it.
func decodeValid[T validator](ctx context.Context, r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}

	if problems := v.Valid(ctx); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}

	return v, nil, nil
}
