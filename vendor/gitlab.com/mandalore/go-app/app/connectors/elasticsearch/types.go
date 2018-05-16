package elasticsearch

import (
	"encoding/json"
	"fmt"
)

// IndexDefinition ...
type IndexDefinition struct {
	Aliases  map[string]json.RawMessage `json:"aliases,omitempty"`
	Mappings map[string]*IndexMappings  `json:"mappings,omitempty"`
	Settings json.RawMessage            `json:"settings,omitempty"`
}

// IdenticalTo ...
func (index *IndexDefinition) IdenticalTo(src *IndexDefinition) error {
	var isDynamic bool
	if _, found := index.Mappings["_default_"]; found {
		isDynamic = true
	}
	for indexType, dstMapping := range index.Mappings {
		srcMapping, found := src.Mappings[indexType]
		if !found {
			return fmt.Errorf("src mapping does not contain the type %s", indexType)
		}
		if found && srcMapping.Meta.Version != dstMapping.Meta.Version {
			return fmt.Errorf("version mismatch [src:%s][dst:%s]", srcMapping.Meta.Version, dstMapping.Meta.Version)
		}
	}
	for indexType := range src.Mappings {
		if _, found := index.Mappings[indexType]; !found && !isDynamic {
			return fmt.Errorf("dst mapping does not contain the type %s", indexType)
		}
	}

	return nil
}

// IndexMappings ...
type IndexMappings struct {
	Meta struct {
		Version string `json:"version,omitempty"`
	} `json:"_meta,omitempty"`
	Properties       json.RawMessage `json:"properties,omitempty"`
	DynamicTemplates json.RawMessage `json:"dynamic_templates,omitempty"`
}

// Response ...
type Response struct {
	ScrollID string            `json:"_scroll_id,omitempty"`
	Took     int               `json:"took,omitempty"`
	Hits     ResponseHits      `json:"hits,omitempty"`
	Errors   bool              `json:"errors,omitempty"`
	Items    []json.RawMessage `json:"items,omitempty"`
}

// ResponseHits ...
type ResponseHits struct {
	Total int           `json:"total,omitempty"`
	Hits  []ResponseHit `json:"hits,omitempty"`
}

// ResponseHit ...
type ResponseHit struct {
	Index   string          `json:"_index,omitempty"`
	Type    string          `json:"_type,omitempty"`
	ID      string          `json:"_id,omitempty"`
	Version int64           `json:"_version,omitempty"`
	Found   bool            `json:"found"`
	Source  json.RawMessage `json:"_source,omitempty"`
}

// BulkResponse contains the whole bulk response as returned by Elasticsearch.
type BulkResponse struct {
	Took   int                `json:"took,omitempty"`
	Errors bool               `json:"errors,omitempty"`
	Items  []BulkItemResponse `json:"items,omitempty"`
}

// BulkItemResponse ...
type BulkItemResponse struct {
	Index *IndexResponse `json:"index,omitempty"`
}

// IndexResponse ...
type IndexResponse struct {
	Index   string          `json:"_index,omitempty"`
	Type    string          `json:"_type,omitempty"`
	ID      string          `json:"_id,omitempty"`
	Version int64           `json:"_version,omitempty"`
	Result  string          `json:"result,omitempty"`
	Shards  *ShardsResponse `json:"_shards,omitempty"`
	Created bool            `json:"created,omitempty"`
	Status  int             `json:"status,omitempty"`
}

// ShardsResponse ...
type ShardsResponse struct {
	Total      int `json:"total,omitempty"`
	Successful int `json:"successful,omitempty"`
	Failed     int `json:"failed,omitempty"`
}

// BulkIndexRequestHeader ...
type BulkIndexRequestHeader struct {
	Index string `json:"_index,omitempty"`
	Type  string `json:"_type,omitempty"`
	ID    string `json:"_id,omitempty"`
	// TODO: parent, routing, etc, etc
}
