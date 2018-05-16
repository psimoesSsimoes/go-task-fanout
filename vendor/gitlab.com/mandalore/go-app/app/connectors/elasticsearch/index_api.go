package elasticsearch

import (
	"encoding/json"
	"fmt"

	"gitlab.com/mandalore/go-app/app"
)

// InstallIndex creates a new index named `<indexName>-v<indexVersion>` and then creates
// an alias named <indexName> for it, removing the alias from any other index. If the
// indexVersion is the zero value string then no alias is created and the index will be
// created using only `<indexName>`.
func (repo *Connector) InstallIndex(indexName, indexVersion string, indexDefinition *IndexDefinition) error {
	var indexFullName string

	if indexVersion != "" {
		indexFullName = fmt.Sprintf("%s-v%s", indexName, indexVersion)
	} else {
		indexFullName = indexName
	}

	status, response, err := repo.Do("GET", indexFullName, nil)
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "unexpected error when retrieving index settings", err)
	}

	switch status {
	case 200:
		currentDefinition := make(map[string]*IndexDefinition)
		if err := json.Unmarshal(response, &currentDefinition); err != nil {
			return app.NewError(app.ErrorUnexpected, "failed to parse response from get index settings", err)
		}

		if err := indexDefinition.IdenticalTo(currentDefinition[indexFullName]); err != nil {
			return err
		}

		if indexVersion != "" {
			if _, found := currentDefinition[indexFullName].Aliases[indexName]; found {
				return nil
			}

			return repo.CreateAlias(indexFullName, indexName)
		}

	case 404:
		reqBody, err := json.Marshal(indexDefinition)
		if err != nil {
			return app.NewError(app.ErrorUnexpected, "failed to marshal index configuration as JSON", err)
		}
		status, response, err = repo.Do("PUT", indexFullName, reqBody)
		if err != nil {
			return app.NewError(app.ErrorUnexpected, "create index failed", err)
		}
		if status != 200 {
			app.Logger.Debugf("response: %s", string(response))
			return app.NewError(app.ErrorUnexpected, fmt.Sprintf("create index failed with unexpected status code [status_code:%d]", status), nil)
		}

		if indexVersion != "" {
			return repo.CreateAlias(indexFullName, indexName)
		}

	default:
		return app.NewError(app.ErrorUnexpected, fmt.Sprintf("unexpected response when fetching index settings [status_code:%d]", status), nil)
	}

	return nil
}

// CreateAlias ...
func (repo *Connector) CreateAlias(indexName, indexAlias string) error {
	var request string

	status, response, err := repo.Do("HEAD", indexAlias, nil)
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "error checking for index", err)
	}

	switch status {
	case 200:
		request = fmt.Sprintf(
			`{"actions":[{"remove":{"index":"%s","alias":"%s"}},{"add":{"index":"%s","alias":"%s"}}]}`,
			indexAlias,
			indexAlias,
			indexName,
			indexAlias,
		)
	case 404:
		request = fmt.Sprintf(
			`{"actions":[{"add":{"index":"%s","alias":"%s"}}]}`,
			indexName,
			indexAlias,
		)
	default:
		return app.NewError(app.ErrorUnexpected, fmt.Sprintf("unexpected status code [status_code:%d][response_body:%s]", status, string(response)), nil)
	}

	app.Logger.Debugf("creating index alias [doc:%s]", request)

	status, response, err = repo.Do("POST", "_aliases", []byte(request))
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "create alias failed", err)
	}
	if status != 200 {
		app.Logger.Debugf("response: %s", string(response))
		return app.NewError(app.ErrorUnexpected, fmt.Sprintf("create alias failed with unexpected status code [status_code:%d][response:%s]", status, string(response)), nil)
	}

	return nil
}
