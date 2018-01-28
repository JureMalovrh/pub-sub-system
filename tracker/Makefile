build:
	@go build -o dist/tracker -i $(shell find cmd -name "*.go")

run: build
	@./dist/tracker -config config.toml

#adds demo data into enviroment
demo-data: demo-data/remove
	@sudo docker-compose exec database /opt/demo/insert_data.sh

demo-data/remove:
	@sudo docker-compose exec database /opt/demo/drop_data.sh

qa:
	go test ./handler/ ./database/ -v -race

help:
	@echo Commands for running and dealing with project
	@echo "\"build\" - builds code"
	@echo "\"run\" - builds code and run project"
	@echo "\"demo-data\" - inserts demo-data in database"
	@echo "\"demo-data/remove\" - removes demo-data from database" 
	@echo "\"qa\" - runs a test for project"