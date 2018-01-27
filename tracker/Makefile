

build:
	@go build -o dist/tracker -i $(shell find cmd -name "*.go")

run: build
	@./dist/tracker -config config.toml

demo-data: demo-data/remove
	@sudo docker-compose exec database /opt/demo/insert_data.sh

demo-data/remove:
	@sudo docker-compose exec database /opt/demo/drop_data.sh

qa:
	go test ./handler/ ./database/ -v -race