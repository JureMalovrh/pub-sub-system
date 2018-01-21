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