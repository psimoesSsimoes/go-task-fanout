package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"gitlab.com/mandalore/go-app/app"
)

// BulkStatistics is a collection of statistical data for one or more bulk requests.
type BulkStatistics struct {
	totalRequests int
	tookMin       int
	tookMax       int
	tookAverage   float64
}

// Bulk is a helper for doing bulk requests.
type Bulk struct {
	conn        *Connector
	body        *bytes.Buffer
	numRequests int
	stats       *BulkStatistics
	bufferSize  int
	mutex       *sync.Mutex
}

// NewBulk ...
func NewBulk(conn *Connector) *Bulk {
	return &Bulk{
		conn:       conn,
		bufferSize: 1000,
		mutex:      new(sync.Mutex),
		stats: &BulkStatistics{
			totalRequests: 0,
			tookMin:       0,
			tookMax:       0,
			tookAverage:   0.0,
		},
	}
}

// Index adds a document index operation to the bulk request.
func (bulk *Bulk) Index(indexName, typeName, id string, record interface{}) error {
	bulk.mutex.Lock()
	defer bulk.mutex.Unlock()

	return bulk.index(indexName, typeName, id, record)
}

// Flush forces a bulk request buffer flush.
func (bulk *Bulk) Flush() (*BulkResponse, error) {
	bulk.mutex.Lock()
	defer bulk.mutex.Unlock()

	return bulk.flush()
}

// IsFull returns true if the bulk request is full.
func (bulk *Bulk) IsFull() bool {
	bulk.mutex.Lock()
	defer bulk.mutex.Unlock()

	return bulk.isFull()
}

// IsEmpty returns true if the bulk request is full.
func (bulk *Bulk) IsEmpty() bool {
	bulk.mutex.Lock()
	defer bulk.mutex.Unlock()

	return bulk.isEmpty()
}

// Stats ...
func (bulk *Bulk) Stats() *BulkStatistics {
	return bulk.stats
}

func (bulk *Bulk) generateHeader(action, indexName, typeName, id string) string {
	switch action {
	case "index":
		if id == "" {
			return fmt.Sprintf(`{"index":{"_index":"%s","_type":"%s"}}`, indexName, typeName)
		}
		return fmt.Sprintf(`{"index":{"_index":"%s","_type":"%s","_id":"%s"}}`, indexName, typeName, id)
	default:
		panic("invalid bulk action provided")
	}
}

func (bulk *Bulk) index(indexName, typeName, id string, record interface{}) (err error) {
	if bulk.numRequests >= bulk.bufferSize {
		return app.NewError(app.ErrorBufferFull, "bulk operation has too many pending requests", nil)
	}

	// write the header
	header := bulk.generateHeader("index", indexName, typeName, id)
	if bulk.body == nil {
		bulk.body = bytes.NewBufferString(header)
	} else {
		if _, err = bulk.body.WriteString(header); err != nil {
			return err
		}
	}
	if _, err = bulk.body.WriteString("\n"); err != nil {
		return err
	}

	// write the body
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal ")
	}
	if _, err = bulk.body.Write(data); err != nil {
		return err
	}
	if _, err = bulk.body.WriteString("\n"); err != nil {
		return err
	}

	bulk.numRequests++

	return nil
}

func (bulk *Bulk) isFull() bool {
	if bulk.numRequests >= bulk.bufferSize {
		return true
	}
	return false
}

func (bulk *Bulk) isEmpty() bool {
	if bulk.numRequests == 0 {
		return true
	}
	return false
}

func (bulk *Bulk) flush() (*BulkResponse, error) {
	var esResponse BulkResponse

	if bulk.body == nil {
		app.Logger.Warnf("bulk request finished without any request to flush")

		return nil, nil
	}

	status, response, err := bulk.conn.Do("POST", "_bulk", bulk.body.Bytes())
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, app.NewError(app.ErrorUnexpected, fmt.Sprintf("failed to run bulk request [status:%d][response.body:%s]", status, string(response)), nil)
	}
	if err := json.Unmarshal(response, &esResponse); err != nil {
		return nil, app.NewError(app.ErrorUnexpected, "failed to parse elasticsearch response", err)
	}

	// just some debug stuff
	if bulk.stats.totalRequests > 0 {
		total := float64(bulk.stats.totalRequests + bulk.numRequests)
		preTotal := float64(bulk.stats.totalRequests) / total
		bulk.stats.tookAverage = bulk.stats.tookAverage*preTotal + float64(esResponse.Took)/total
	} else {
		bulk.stats.tookAverage = float64(esResponse.Took) / float64(bulk.numRequests)
	}

	// reset everything
	bulk.body.Reset()
	bulk.numRequests = 0
	bulk.stats.totalRequests += bulk.numRequests
	if bulk.stats.tookMin == 0 || bulk.stats.tookMin > esResponse.Took {
		bulk.stats.tookMin = esResponse.Took
	}
	if bulk.stats.tookMax < esResponse.Took {
		bulk.stats.tookMax = esResponse.Took
	}

	return &esResponse, nil
}

// BulkSimple is similar to Bulk except it handles errors in a very generic and simple way.
// Less powerful but more useful when error handling per operation is not a requirement.
// It also uses auto-flush when the bulk buffer is full.
type BulkSimple struct {
	Bulk
}

// NewBulkSimple ...
func NewBulkSimple(conn *Connector) *BulkSimple {
	return &BulkSimple{
		Bulk: Bulk{
			conn:       conn,
			bufferSize: 1000,
			mutex:      new(sync.Mutex),
			stats: &BulkStatistics{
				totalRequests: 0,
				tookMin:       0,
				tookMax:       0,
				tookAverage:   0.0,
			},
		},
	}
}

// Index adds a document index operation to the bulk request.
func (bulk *BulkSimple) Index(indexName, typeName, id string, record interface{}) error {
	bulk.mutex.Lock()
	defer bulk.mutex.Unlock()

	if err := bulk.index(indexName, typeName, id, record); err != nil {
		return err
	}

	// REVIEW: consider using bulk.Len() as a secondary (or even primary) flush trigger
	// is it flush time?
	bulk.numRequests++
	if bulk.numRequests >= bulk.bufferSize {
		if _, err := bulk.flush(); err != nil {
			return err
		}
	}

	return nil
}

// Flush forces a bulk request buffer flush.
func (bulk *BulkSimple) Flush() error {
	bulk.mutex.Lock()
	defer bulk.mutex.Unlock()

	response, err := bulk.flush()
	if err != nil {
		return err
	}

	if response.Errors {
		return app.NewError(app.ErrorUnexpected, "bulk request failed", nil)
	}

	return nil
}
