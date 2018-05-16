// +build integration

package elasticsearch

import (
	"encoding/json"
	"testing"

	"gitlab.com/mandalore/go-app/app"
)

func TestIndexCreation(t *testing.T) {
	cfg := &Config{
		BaseURI: "http://localhost:9200",
		Indices: map[string]*IndexConfig{
			"test-01": &IndexConfig{
				IndexName:    "test-01",
				IndexVersion: "1",
			},
		},
	}

	conn := NewConnector(cfg)

	var sampleIndexMapping IndexDefinition
	var data []byte
	data, err := app.Util.LoadFile("testdata/sample-schema-01.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &sampleIndexMapping); err != nil {
		t.Fatal(err)
	}
	if err = conn.InstallIndex(cfg.Indices["test-01"].IndexName, cfg.Indices["test-01"].IndexVersion, &sampleIndexMapping); err != nil {
		t.Fatal(err)
	}

	// TODO: validate that the index is correctly indexed
}
