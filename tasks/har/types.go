package har

// Document represents the root of a HAR file.
type Document struct {
	Log Log `json:"log"`
}

// Log holds the HAR log object details we care about.
type Log struct {
	Version string   `json:"version"`
	Entries []Entry  `json:"entries"`
	Pages   []Page   `json:"pages"`
	Creator *Creator `json:"creator"`
}

// Page contains metadata about captured browser pages.
type Page struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	StartedDateTime string       `json:"startedDateTime"`
	PageTimings     *PageTimings `json:"pageTimings"`
}

// PageTimings captures load timing metrics.
type PageTimings struct {
	OnContentLoad float64 `json:"onContentLoad"`
	OnLoad        float64 `json:"onLoad"`
}

// Creator describes the HAR generator.
type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Entry describes a single HTTP exchange.
type Entry struct {
	Pageref         string   `json:"pageref"`
	StartedDateTime string   `json:"startedDateTime"`
	Time            float64  `json:"time"`
	Request         Request  `json:"request"`
	Response        Response `json:"response"`
}

// Request captures outbound request details.
type Request struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []NameValue `json:"headers"`
	QueryString []NameValue `json:"queryString"`
	PostData    *PostData   `json:"postData"`
	Cookies     []NameValue `json:"cookies"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
}

// PostData captures request payloads.
type PostData struct {
	MimeType string      `json:"mimeType"`
	Text     string      `json:"text"`
	Params   []NameValue `json:"params"`
}

// Response captures inbound response details.
type Response struct {
	Status       int         `json:"status"`
	StatusText   string      `json:"statusText"`
	HTTPVersion  string      `json:"httpVersion"`
	Headers      []NameValue `json:"headers"`
	Cookies      []NameValue `json:"cookies"`
	Content      Content     `json:"content"`
	RedirectURL  string      `json:"redirectURL"`
	HeadersSize  int         `json:"headersSize"`
	BodySize     int         `json:"bodySize"`
	TransferSize int         `json:"_transferSize"`
	FetchedViaSW bool        `json:"_fetchedViaServiceWorker"`
}

// Content describes the response body.
type Content struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// NameValue is a reusable pair type for HAR metadata.
type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
