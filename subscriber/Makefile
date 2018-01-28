build:
	@go build -o dist/client -i $(shell find cmd -name "*.go")

run/aggregator: build
	@./dist/client -agg=true

run/aggregator/%: build
	@./dist/client -agg=true -filter=$*

run/printer: build
	@./dist/client

run/printer/%: build
	@./dist/client -filter=$*

qa:
	go test -v -race -timeout 30s ./cmd

help:
	@echo Commands for running and dealing with project
	@echo "\"build\" - builds code"
	@echo "\"run/aggregator\" - runs service as an aggregator"
	@echo "\"run/aggregator/ID\" - runs service as an aggregator, with filter being ID"
	@echo "\"run/printer\" - runs service as an printer"
	@echo "\"run/printer/ID\" - runs service as an printer, with filter being ID"
	@echo "\"qa\" - runs tests for this service"
