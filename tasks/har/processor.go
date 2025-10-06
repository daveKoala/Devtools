package har

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Processor converts HAR documents into OpenAPI definitions.
type Processor struct {
	Domains     []string
	headerRules map[string]HeaderRule
}

// NewProcessor builds a Processor configured with allowed domains.
func NewProcessor(domains []string) *Processor {
	cleaned := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}
		cleaned = append(cleaned, domain)
	}
	return &Processor{Domains: cleaned, headerRules: make(map[string]HeaderRule)}
}

// HeaderRule controls how request headers should appear in generated specs.
type HeaderRule struct {
	Name        string
	Replacement string
	Required    bool
	Description string
}

// AllowHeader registers a header that should be surfaced in the OpenAPI document.
// If replacement is provided it will be used as the example value to avoid
// leaking sensitive tokens extracted from the HAR file.
func (p *Processor) AllowHeader(name, replacement string) {
	if p.headerRules == nil {
		p.headerRules = make(map[string]HeaderRule)
	}

	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return
	}

	// Preserve the original casing provided by the caller for display purposes.
	p.headerRules[key] = HeaderRule{
		Name:        name,
		Replacement: replacement,
	}
}

// GenerateOpenAPIDocument produces an OpenAPI document JSON payload from the HAR.
func (p *Processor) GenerateOpenAPIDocument(ctx context.Context, doc *Document) ([]byte, *OpenAPIDocument, error) {
	if doc == nil {
		return nil, nil, fmt.Errorf("nil HAR document provided")
	}

	paths := make(map[string]*PathItem)
	allowedHosts := make(map[string]struct{})
	for _, entry := range doc.Log.Entries {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}

		reqURL, ok := p.acceptEntry(entry)
		if !ok {
			continue
		}

		host := strings.ToLower(strings.TrimSpace(reqURL.Hostname()))
		if host != "" {
			allowedHosts[host] = struct{}{}
		}

		cleanPath := reqURL.EscapedPath()
		if cleanPath == "" {
			cleanPath = "/"
		}

		if strings.HasSuffix(cleanPath, "/") && len(cleanPath) > 1 {
			cleanPath = strings.TrimRight(cleanPath, "/")
		}

		item := paths[cleanPath]
		if item == nil {
			item = &PathItem{}
			paths[cleanPath] = item
		}

		operation := p.buildOperation(entry)

		switch strings.ToUpper(entry.Request.Method) {
		case "GET":
			item.Get = operation
		case "POST":
			item.Post = operation
		case "PUT":
			item.Put = operation
		case "PATCH":
			item.Patch = operation
		case "DELETE":
			item.Delete = operation
		case "HEAD":
			item.Head = operation
		case "OPTIONS":
			item.Options = operation
		case "TRACE":
			item.Trace = operation
		default:
			// unsupported method, skip for now
		}
	}

	titles := collectPageTitles(doc)
	infoTitle := deriveTitle(titles, p.Domains, allowedHosts)

	opDoc := &OpenAPIDocument{
		OpenAPI: "3.1.0",
		Info: OpenAPIInfo{
			Title:   infoTitle,
			Version: "0.1.0",
			Summary: fmt.Sprintf("Generated from %d HAR entries", len(doc.Log.Entries)),
		},
		Paths: paths,
	}

	encoded, err := json.MarshalIndent(opDoc, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("encode OpenAPI JSON: %w", err)
	}

	return encoded, opDoc, nil
}

func (p *Processor) acceptEntry(entry Entry) (*url.URL, bool) {
	reqURL, err := url.Parse(entry.Request.URL)
	if err != nil {
		return nil, false
	}

	host := strings.ToLower(strings.TrimSpace(reqURL.Hostname()))
	if host == "" {
		return nil, false
	}

	if isBlockedDomain(host) {
		return nil, false
	}

	if p.allowHost(host) {
		return reqURL, true
	}

	if p.matchHeaderDomains(entry.Request.Headers) {
		return reqURL, true
	}

	return nil, false
}

func (p *Processor) allowHost(host string) bool {
	if len(p.Domains) == 0 {
		return false
	}

	for _, domain := range p.Domains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

func (p *Processor) matchHeaderDomains(headers []NameValue) bool {
	if len(headers) == 0 {
		return false
	}

	for _, header := range headers {
		name := strings.ToLower(strings.TrimSpace(header.Name))
		switch name {
		case "origin", "referer":
			host := hostFromURL(header.Value)
			if host == "" {
				continue
			}
			if isBlockedDomain(host) {
				continue
			}
			if p.allowHost(host) {
				return true
			}
		}
	}

	return false
}

func hostFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		clean := strings.ToLower(raw)
		clean = strings.Trim(clean, "/")
		return clean
	}

	return strings.ToLower(strings.TrimSpace(parsed.Hostname()))
}

func (p *Processor) buildOperation(entry Entry) *Operation {
	req := entry.Request
	res := entry.Response

	summary := fmt.Sprintf("%s %s", req.Method, trimURL(req.URL))
	desc := fmt.Sprintf("Captured response status %d %s", res.Status, res.StatusText)

	operationID := strings.ToLower(req.Method) + "_" + sanitiseOperationID(req.URL)

	responses := map[string]*OpenAPIResponse{}
	statusKey := strconv.Itoa(res.Status)
	if res.Status == 0 {
		statusKey = "default"
	}

	responses[statusKey] = p.buildResponse(res, desc)

	operation := &Operation{
		OperationID: operationID,
		Summary:     summary,
		Description: desc,
		Responses:   responses,
	}

	if len(p.headerRules) > 0 {
		operation.Parameters = append(operation.Parameters, p.buildHeaderParameters(entry.Request.Headers)...)
	}

	if body := p.buildRequestBody(req); body != nil {
		operation.RequestBody = body
	}

	return operation
}

func (p *Processor) buildResponse(res Response, desc string) *OpenAPIResponse {
	mediaType := strings.TrimSpace(res.Content.MimeType)
	content := map[string]*MediaTypeDefinition{}

	if mediaType != "" {
		example := exampleFromBody(mediaType, res.Content.Text)
		content[mediaType] = &MediaTypeDefinition{
			Schema:  &Schema{Type: inferTypeFromMime(mediaType)},
			Example: example,
		}
	}

	return &OpenAPIResponse{
		Description: strings.TrimSpace(desc),
		Content:     content,
	}
}

func (p *Processor) buildRequestBody(req Request) *RequestBody {
	if req.PostData == nil {
		return nil
	}

	mime := strings.TrimSpace(req.PostData.MimeType)
	bodyText := strings.TrimSpace(req.PostData.Text)

	if mime == "" {
		if len(req.PostData.Params) > 0 {
			mime = "application/x-www-form-urlencoded"
		} else if looksLikeJSON(bodyText) {
			mime = "application/json"
		} else {
			mime = "text/plain"
		}
	}

	if bodyText == "" && len(req.PostData.Params) > 0 {
		pairs := make([]string, 0, len(req.PostData.Params))
		for _, param := range req.PostData.Params {
			pairs = append(pairs, fmt.Sprintf("%s=%s", param.Name, param.Value))
		}
		bodyText = strings.Join(pairs, "&")
	}

	example := exampleFromBody(mime, bodyText)
	if example == nil {
		return nil
	}

	content := map[string]*MediaTypeDefinition{
		mime: {
			Schema:  &Schema{Type: inferTypeFromMime(mime)},
			Example: example,
		},
	}

	required := strings.EqualFold(req.Method, "POST") || strings.EqualFold(req.Method, "PUT") || strings.EqualFold(req.Method, "PATCH")

	return &RequestBody{
		Required: required,
		Content:  content,
	}
}

func (p *Processor) buildHeaderParameters(headers []NameValue) []*Parameter {
	if len(headers) == 0 || len(p.headerRules) == 0 {
		return nil
	}

	params := make([]*Parameter, 0, len(p.headerRules))

	for _, header := range headers {
		key := strings.ToLower(strings.TrimSpace(header.Name))
		rule, ok := p.headerRules[key]
		if !ok {
			continue
		}

		value := strings.TrimSpace(header.Value)
		if rule.Replacement != "" {
			value = rule.Replacement
		}

		param := &Parameter{
			Name:        rule.Name,
			In:          "header",
			Required:    rule.Required,
			Description: rule.Description,
			Schema: &Schema{
				Type: "string",
			},
		}

		if value != "" {
			param.Example = value
		}

		params = append(params, param)
	}

	return params
}

func exampleFromBody(mime, text string) interface{} {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	if strings.Contains(strings.ToLower(mime), "json") || looksLikeJSON(trimmed) {
		var payload interface{}
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			return payload
		}
	}

	return truncateString(trimmed, 2048)
}

func looksLikeJSON(body string) bool {
	if body == "" {
		return false
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(body)
	last, _ := utf8.DecodeLastRuneInString(body)
	return (first == '{' && last == '}') || (first == '[' && last == ']')
}

func truncateString(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "…"
}

func trimURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	sanitised := path.Clean(parsed.Path)
	if sanitised == "." {
		sanitised = "/"
	}
	if parsed.RawQuery != "" {
		sanitised += "?" + parsed.RawQuery
	}
	return sanitised
}

func sanitiseOperationID(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "unknown"
	}

	id := strings.Trim(parsed.Path, "/")
	if id == "" {
		id = "root"
	}

	replacer := strings.NewReplacer("/", "_", "-", "_", "{", "_", "}", "_")
	id = replacer.Replace(id)
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 32
		case r >= '0' && r <= '9':
			return r
		case r == '_':
			return r
		default:
			return -1
		}
	}, id)
}

func inferTypeFromMime(mime string) string {
	lower := strings.ToLower(mime)
	switch {
	case strings.Contains(lower, "json"):
		return "object"
	case strings.Contains(lower, "xml"):
		return "object"
	case strings.Contains(lower, "html"):
		return "string"
	case strings.Contains(lower, "text"):
		return "string"
	default:
		return "string"
	}
}

func collectPageTitles(doc *Document) []string {
	if doc == nil {
		return nil
	}
	titles := make([]string, 0, len(doc.Log.Pages))
	for _, page := range doc.Log.Pages {
		title := strings.TrimSpace(page.Title)
		if title == "" {
			continue
		}
		titles = append(titles, title)
	}
	sort.Strings(titles)
	return titles
}

func deriveTitle(titles []string, allowedDomains []string, hosts map[string]struct{}) string {
	orderedHosts := orderedHostList(allowedDomains, hosts)
	if len(orderedHosts) > 0 {
		return fmt.Sprintf("%s API", orderedHosts[0])
	}

	if len(titles) == 0 {
		return "Generated API"
	}

	if len(titles) == 1 {
		return titles[0]
	}

	return titles[0] + " et al."
}

func orderedHostList(allowedDomains []string, hosts map[string]struct{}) []string {
	if len(hosts) == 0 {
		return nil
	}

	// Honour the user-specified order of domains when possible.
	result := make([]string, 0, len(hosts))
	seen := make(map[string]struct{}, len(hosts))

	for _, domain := range allowedDomains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		if _, ok := hosts[domain]; ok {
			result = append(result, domain)
			seen[domain] = struct{}{}
		}
	}

	// Append any additional hosts discovered in the HAR that weren't explicitly listed.
	for host := range hosts {
		if _, ok := seen[host]; ok {
			continue
		}
		result = append(result, host)
	}

	sort.Strings(result[len(seen):])
	return result
}
