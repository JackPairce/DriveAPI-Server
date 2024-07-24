include .env
export

build:
	@go build -o ./bin/server

run:
	@./bin/server

dev: build run
