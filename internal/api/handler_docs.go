package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/api/openapi"
)

func (s *Server) ServeOpenAPISpec(ctx context.Context, r gen.ServeOpenAPISpecRequestObject) (gen.ServeOpenAPISpecResponseObject, error) {
	specJSON, err := gen.GetSpecJSON()
	if err != nil {
		log.Printf("500 at ServeOpenAPISpec, GetSpecJSON: %v", err)
		return nil, err
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		log.Printf("500 at ServeOpenAPISpec, Unmarshal: %v", err)
		return nil, err
	}

	return gen.ServeOpenAPISpec200JSONResponse(spec), nil
}

func (s *Server) ServeOpenAPIDocs(ctx context.Context, r gen.ServeOpenAPIDocsRequestObject) (gen.ServeOpenAPIDocsResponseObject, error) {
	data, err := openapi.FS.ReadFile("docs.html")
	if err != nil {
		log.Printf("500 at ServeOpenAPIDocs: %v", err)
		return nil, err
	}

	return gen.ServeOpenAPIDocs200TexthtmlResponse{
		Body:          bytes.NewReader(data),
		ContentLength: int64(len(data)),
	}, nil
}
