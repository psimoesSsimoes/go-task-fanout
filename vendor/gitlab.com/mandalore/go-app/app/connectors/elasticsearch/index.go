package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"gitlab.com/mandalore/go-app/app"
)

// IndexConfig for unmarshalling elasticsearch configurations compatible with JSON and viper's mapstructure.
type IndexConfig struct {
	IndexName    string `json:"name" mapstructure:"name"`
	IndexVersion string `json:"version" mapstructure:"version"`
}

func (config *IndexConfig) isValid() bool {
	// TODO: proper validation (consider using a validation package)
	if config.IndexName == "" || config.IndexVersion == "" {
		return false
	}

	return true
}

// Index ...
type Index struct {
	name    string
	alias   string
	version string
	conn    *Connector
}

// NewIndex ...
func NewIndex(connector *Connector, config *IndexConfig) (*Index, error) {
	if config == nil || !config.isValid() || connector == nil {
		return nil, app.NewError(app.ErrorInvalidParameter, "invalid parameters for index instance", nil)
	}

	return &Index{
		name:    fmt.Sprintf("%s-v%s", config.IndexName, config.IndexVersion),
		version: config.IndexVersion,
		alias:   config.IndexName,
		conn:    connector,
	}, nil
}

// GetAlias returns the index alias.
func (index *Index) GetAlias() string {
	return index.alias
}

// GetName returns the full index name, including version sufix.
func (index *Index) GetName() string {
	return index.name
}

// Install creates the index using the provided configuration.
func (index *Index) Install(definition *IndexDefinition) error {
	return index.conn.InstallIndex(index.alias, index.version, definition)
}

// Do runs a raw request against theindex with the provided parameters.
func (index *Index) Do(method, endpoint string, request []byte) (int, []byte, error) {
	return index.conn.Do(method, fmt.Sprintf("%s/%s", index.name, endpoint), request)
}

func (index *Index) generateEndpoint(docType, id, version string) string {
	endpointBuf := bytes.NewBuffer(make([]byte, len(docType)+len(id)+2+len(version)+31))
	endpointBuf.WriteString("/")
	endpointBuf.WriteString(docType)
	if id != "" {
		endpointBuf.WriteString("/")
		endpointBuf.WriteString(id)
	}
	if version != "" {
		endpointBuf.WriteString("?version_type=external")
		endpointBuf.WriteString("&version=")
		endpointBuf.WriteString(version)
	}

	return endpointBuf.String()
}

// SetRefreshInterval ...
func (index *Index) SetRefreshInterval(value int) (err error) {
	const templateRefreshInterval = `{"settings":{"refresh_interval":"%s"}}`
	var body string

	if value == 0 || value > 300 || value < -1 {
		panic("invalid value for refresh interval")
	}

	if value < 0 {
		body = fmt.Sprintf(templateRefreshInterval, strconv.Itoa(value))
	} else {
		body = fmt.Sprintf(templateRefreshInterval, strconv.Itoa(value)+"s")
	}

	_, _, err = index.Do("PUT", "/_settings", []byte(body))
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "failed to change index refresh interval", err)
	}

	return nil
}

// Set stores record, which must be JSON serializable, in the index under the provided document type and id.
func (index *Index) Set(docType, id string, record interface{}) error {
	return index.Index(docType, id, record)
}

// Get fetches a document and unserializes (JSON) the contents into record.
func (index *Index) Get(docType, id string, record interface{}) error {
	var err error
	var raw json.RawMessage

	raw, err = index.GetDocument(docType, id)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, record)
	if err != nil {
		return err
	}

	return nil
}

// Index ...
func (index *Index) Index(docType, id string, record interface{}) error {
	// if index.conn.bulk != nil {
	// 	return index.bulk.Index(index.indexName, docType, id, record)
	// }

	return index.indexDocument(docType, id, record)
}

func (index *Index) indexDocument(docType, id string, record interface{}) error {
	var endpoint string

	// index.Do("POST", docType
	if id != "" {
		endpoint = fmt.Sprintf("%s/%s", docType, id)
	} else {
		endpoint = docType
	}

	body, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal data")
	}

	status, rawResponse, err := index.Do("POST", endpoint, body)
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "failed to index document", err)
	}

	switch status {
	case 200, 204:
		response := &IndexResponse{}
		if err := json.Unmarshal(rawResponse, response); err != nil {
			return app.NewError(app.ErrorUnexpected, "unexpected error unmarshalling elasticsearch response", err)
		}
		return nil
		// return response, nil
	case 409:
		return app.NewError(app.ErrorConflict, fmt.Sprintf("document already exists"), nil)
	default:
		return app.NewError(app.ErrorUnexpected, fmt.Sprintf("unexpected response [code:%d][response.body:%s]", status, rawResponse), nil)
	}
}

// GetDocument returns a single document.
func (index *Index) GetDocument(docType, id string) (json.RawMessage, app.Error) {
	endpoint := index.generateEndpoint(docType, id, "")

	status, response, err := index.conn.Do("GET", endpoint, nil)
	if err != nil {
		return nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("request failed [url:%s]", endpoint), err)
	}
	switch status {
	case 200:
		doc := &ResponseHit{}
		if err := json.Unmarshal(response, doc); err != nil {
			return nil, app.NewError(app.ErrorUnexpected, "error parsing Elasticsearch response", err)
		}
		return doc.Source, nil
	case 404:
		return nil, nil
	default:
		return nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("invalid response from index [code:%d][response.body:%s]", status, response), nil)
	}
}

// Search runs a search using
func (index *Index) Search(docType, query string) (*Response, error) {
	var esRes Response
	var endpoint string

	if docType != "" {
		endpoint = fmt.Sprintf("%s/%s/_search", index.name, docType)
	} else {
		endpoint = fmt.Sprintf("%s/_search", index.name)
	}

	status, response, err := index.conn.Do("POST", endpoint, []byte(query))
	if err != nil {
		return nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("unexpected error when executing search"), err)
	}

	switch status {
	case 200:
		if err := json.Unmarshal(response, &esRes); err != nil {
			return nil, app.NewError(app.ErrorUnexpected, "error unmarshalling Elasticsearch response", err)
		}
		return &esRes, nil
	default:
		return nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("unexpected response when executing search [code:%d][message:%s]", status, string(response)), nil)
	}

}
