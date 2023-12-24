test:
	go test -count 1 -timeout 30s -run ^Test ./...
	go test -count 1 -timeout 30s -run ^Test github.com/lxzan/memorycache/benchmark

bench:
	go test -benchmem -run=^$$ -bench . github.com/lxzan/memorycache/benchmark

cover:
	go test -coverprofile=./bin/cover.out --cover ./...