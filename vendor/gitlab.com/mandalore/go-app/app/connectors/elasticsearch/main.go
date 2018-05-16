package elasticsearch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/mandalore/go-app/app"
)

// Config for unmarshalling elasticsearch configurations compatible with JSON and viper's mapstructure.
type Config struct {
	BaseURI string                  `json:"base_uri" mapstructure:"base_uri"`
	Indices map[string]*IndexConfig `json:"indices"`
	// IndexAlias at the root level is deprecated
	IndexAlias string `json:"index_alias" mapstructure:"index_alias"`
	// IndexVersion at the root level is deprecated
	IndexVersion string `json:"index_version" mapstructure:"index_version"`
}

func (config *Config) isValid() bool {
	// TODO: proper validation (consider using a validation package)
	if config.BaseURI == "" {
		return false
	}
	for _, indexConfig := range config.Indices {
		if indexConfig.IndexName == "" || indexConfig.IndexVersion == "" {
			return false
		}
	}
	if config.IndexAlias != "" || config.IndexVersion != "" {
		return false
	}

	return true
}

func (config *Config) validate() error {
	// TODO: proper validation (consider using a validation package)
	// TODO: put all errors in a slice
	if config.BaseURI == "" {
		return app.NewError(app.ErrorInvalidParameter, "configuration has empty value for .base_uri", nil)
	}
	for indexName, indexConfig := range config.Indices {
		if indexConfig.IndexName == "" {
			return app.NewError(app.ErrorInvalidParameter, fmt.Sprintf("configuration has empty value for .indices.%s.name", indexName), nil)
		}
		if indexConfig.IndexVersion == "" {
			return app.NewError(app.ErrorInvalidParameter, fmt.Sprintf("configuration has empty value for .indices.%s.version", indexName), nil)
		}
	}
	if config.IndexAlias != "" {
		return app.NewError(app.ErrorInvalidParameter, "configuration .index_alias has been deprecated", nil)
	}
	if config.IndexVersion != "" {
		return app.NewError(app.ErrorInvalidParameter, "configuration .index_version has been deprecated", nil)
	}

	return nil
}

// Connector ...
type Connector struct {
	client          *http.Client
	bulk            *BulkSimple
	baseURI         string
	maxBulkRequests int
	asyncIndexing   bool
}

// NewConnector ...
func NewConnector(config *Config) *Connector {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 50

	if !config.isValid() {
		panic("invalid configuration")
	}

	rep := &Connector{
		client:          &http.Client{},
		baseURI:         config.BaseURI,
		maxBulkRequests: 1000,
		asyncIndexing:   false,
	}

	return rep
}

// SetBulkMaxRequests sets the max number of buffered requests in a bulk session.
func (repo *Connector) SetBulkMaxRequests(max int) {
	repo.maxBulkRequests = max
}

// StartBulk begins a bulk request.
func (repo *Connector) StartBulk() error {
	if repo.bulk != nil {
		return fmt.Errorf("bulk request already started")
	}

	repo.bulk = NewBulkSimple(repo)

	return nil
}

// EndBulk terminates a bulk request, issuing any remaining requests.
func (repo *Connector) EndBulk() error {
	if repo.bulk == nil {
		return fmt.Errorf("no active bulk request")
	}

	err := repo.bulk.Flush()
	repo.bulk = nil

	return err
}

// Do runs a raw request against elasticsearch with the provided parameters.
func (repo *Connector) Do(method, endpoint string, request []byte) (int, []byte, error) {
	var uri string

	if endpoint != "" {
		uri = fmt.Sprintf("%s/%s", repo.baseURI, endpoint)
	} else {
		uri = repo.baseURI
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(request))
	if err != nil {
		return 0, nil, err
	}
	if request != nil {
		req.Header.Add("ContentType", "application/json")
	}

	res, err := repo.client.Do(req)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}

		return 0, nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("failed to execute request [method:%s][uri:%s]", method, uri), err)
	}

	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return 0, nil, err
	}
	return res.StatusCode, response, nil
}

// DoFromBuffer runs a raw request against elasticsearch with the provided parameters.
func (repo *Connector) DoFromBuffer(method, endpoint string, request *bytes.Buffer) (int, []byte, error) {
	var uri string

	if endpoint != "" {
		uri = fmt.Sprintf("%s/%s", repo.baseURI, endpoint)
	} else {
		uri = repo.baseURI
	}

	req, err := http.NewRequest(method, uri, request)
	if err != nil {
		return 0, nil, err
	}
	if request != nil {
		req.Header.Add("ContentType", "application/json")
	}

	res, err := repo.client.Do(req)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}

		return 0, nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("failed to execute request [method:%s][uri:%s]", method, uri), err)
	}

	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return 0, nil, err
	}
	return res.StatusCode, response, nil
}
