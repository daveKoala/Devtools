package har

// OpenAPIDocument models the minimal data we emit for the generated spec.
type OpenAPIDocument struct {
	OpenAPI string               `json:"openapi"`
	Info    OpenAPIInfo          `json:"info"`
	Paths   map[string]*PathItem `json:"paths"`
}

// OpenAPIInfo describes the API metadata.
type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Summary string `json:"summary,omitempty"`
}

// PathItem captures HTTP operations available on a resource.
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`
}

// Operation represents an HTTP method definition.
type Operation struct {
	OperationID string                      `json:"operationId,omitempty"`
	Summary     string                      `json:"summary,omitempty"`
	Description string                      `json:"description,omitempty"`
	Responses   map[string]*OpenAPIResponse `json:"responses"`
	Parameters  []*Parameter                `json:"parameters,omitempty"`
	RequestBody *RequestBody                `json:"requestBody,omitempty"`
}

// OpenAPIResponse describes a single response object.
type OpenAPIResponse struct {
	Description string                          `json:"description"`
	Content     map[string]*MediaTypeDefinition `json:"content,omitempty"`
}

// Parameter represents an operation-level parameter (e.g. header).
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// MediaTypeDefinition details MIME-specific response payload information.
type MediaTypeDefinition struct {
	Schema  *Schema     `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}

// Schema is a pared-down schema representation.
type Schema struct {
	Type string `json:"type,omitempty"`
}

// RequestBody represents the payload expected by an operation.
type RequestBody struct {
	Description string                          `json:"description,omitempty"`
	Required    bool                            `json:"required,omitempty"`
	Content     map[string]*MediaTypeDefinition `json:"content"`
}
