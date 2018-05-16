.PHONY: test build proto

test:
	go test -v -race ./...

bench:
	go test -v -tags bench -bench Bench ./...

itest:
	go test -v -tags integration ./test/...

coverage:
	./sbin/do.sh coverage $(filter-out $@,$(MAKECMDGOALS))

%: 
	@echo "not a valid target"
	@:
