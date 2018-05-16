TESTS=$$(go list ./... | grep -v /vendor/ | grep -v /tests | sort)

install:
	@echo "=== Installing dependencies ==="
	@dep ensure
	@echo "Done"

test: mocks
	@echo "=== Running tests ==="
	go test -cover ${TESTS}

local-env:
	@echo "=== Starting local services ==="
	docker-compose up

integration-tests:
	@echo "=== Running integration tests ==="
	go test -tags=integration -cover ${TESTS}

mocks:
	CGO_ENABLED=0 $(GOPATH)/bin/mockery -all -dir interactors -outpkg mocks -output tests/mocks/interactors
	CGO_ENABLED=0 $(GOPATH)/bin/mockery -all -dir controllers -outpkg mocks -output tests/mocks/controllers

# build: install
# 	go build -a -installsuffix cgo -o bin/go-project-inbound cmd/api/main.go

