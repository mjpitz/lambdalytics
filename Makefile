.PHONY: build clean deploy

build: bin/collector bin/recorder

bin/collector: cmd/collector/ internal/
	GOOS=linux go build -ldflags="-s -w" -o bin/collector cmd/collector/main.go

bin/recorder: cmd/recorder/ internal/
	GOOS=linux go build -ldflags="-s -w" -o bin/recorder cmd/recorder/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
