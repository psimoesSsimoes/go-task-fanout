bench:
	go test -v -bench 'Bench' -benchmem .

test:
	go test -v -race .

bench-profile:
	mkdir -p ./.tmp
	go test -v ./.tmp/cpu.pprof -bench 'Bench' -benchmem .
	go tool pprof ./.tmp/cpu.pprof

coverage:
	go test -coverprofile ./coverage.out
	go tool cover -html=./coverage.out -o ./coverage.html
	rm ./coverage.out
