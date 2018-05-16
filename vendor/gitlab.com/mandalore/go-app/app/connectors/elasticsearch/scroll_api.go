package elasticsearch

import (
	"encoding/json"
	"fmt"

	"gitlab.com/mandalore/go-app/app"
)

// NewScroll initializes an elasticsearch scroll search across all documents of an index in no specific order.
// The scrollCacheMinutes specifies the time, in minutes, to keep the scroll search in Elasticsearch.
// More info at https://www.elastic.co/guide/en/elasticsearch/reference/current/search-request-scroll.html.
func (repo *Connector) NewScroll(indexName string, docType string, limit int, scrollCacheMinutes int) *Scroll {
	return &Scroll{
		indexName: indexName,
		docType:   docType,
		limit:     limit,
		time:      fmt.Sprintf("%dm", scrollCacheMinutes),
		repo:      repo,
	}
}

// Scroll represents a scroll search in elasticsearch. Use the RawStorage's StartScroll method for initializing a scroll search.
type Scroll struct {
	indexName       string
	docType         string
	limit           int
	time            string
	id              string
	repo            *Connector
	started         bool
	currentResponse *Response
}

// Next iterates to next iteration and returns if there is more iterations
func (scroll *Scroll) Next() (bool, app.Error) {
	var err app.Error
	scroll.currentResponse, err = scroll.Iterate()
	if err != nil {
		return false, err
	}

	if scroll.currentResponse == nil || len(scroll.currentResponse.Hits.Hits) <= 0 {
		return false, nil
	}

	return true, nil
}

// Scan unmarshal all hits sources to desc
func (scroll *Scroll) Scan(desc interface{}) app.Error {
	if scroll.currentResponse == nil {
		return app.NewError(app.ErrorDevPoo, "nothing to scan", nil)
	}

	rawDocs := []json.RawMessage{}

	for _, hit := range scroll.currentResponse.Hits.Hits {
		rawDocs = append(rawDocs, hit.Source)
	}

	raw, err := json.Marshal(rawDocs)
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "failed to marshal response", err)
	}

	if err := json.Unmarshal(raw, desc); err != nil {
		return app.NewError(app.ErrorUnexpected, "failed to unmarshal elasticsearch response", err)
	}

	return nil
}

// Iterate runs the next scroll iteration.
func (scroll *Scroll) Iterate() (*Response, app.Error) {
	if !scroll.started {
		return scroll.start()
	}

	status, rawResponse, err := scroll.repo.Do("POST", "_search/scroll", []byte(fmt.Sprintf(`{"scroll":"%s","scroll_id":"%s"}`, scroll.time, scroll.id)))
	if err != nil {
		return nil, app.NewUnexpectedError("failed to iterate scroll", err)
	}
	if status != 200 {
		return nil, app.NewErrorf(app.ErrorUnexpected, nil, "failed to iterate scroll, unexpected response from elasticsearch [response:%s]", string(rawResponse))
	}

	elasticResponse := &Response{}

	if err := json.Unmarshal(rawResponse, elasticResponse); err != nil {
		return nil, app.NewError(app.ErrorUnexpected, "failed to iterate scroll, unmarshal response error", err)
	}

	scroll.id = elasticResponse.ScrollID

	return elasticResponse, nil
}

func (scroll *Scroll) start() (*Response, app.Error) {
	uri := fmt.Sprintf("%s/%s/_search?scroll=%s", scroll.indexName, scroll.docType, scroll.time)

	status, rawResponse, err := scroll.repo.Do("POST", uri, []byte(fmt.Sprintf(`{"size":%d,"sort":["_doc"]}`, scroll.limit)))
	if err != nil {
		return nil, app.NewUnexpectedError("failed to initialize scroll", err)
	}
	if status != 200 {
		return nil, app.NewErrorf(app.ErrorUnexpected, nil, "failed to initialize scroll, unexpected response from elasticsearch [response:%s]", string(rawResponse))
	}

	elasticResponse := &Response{}

	if err := json.Unmarshal(rawResponse, elasticResponse); err != nil {
		return nil, app.NewError(app.ErrorUnexpected, "failed to unmarshal elasticsearch response", err)
	}

	scroll.id = elasticResponse.ScrollID
	scroll.started = true

	return elasticResponse, nil
}
