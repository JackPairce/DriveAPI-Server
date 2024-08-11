include .env
export

build:
	@go build -o ./bin/server

run:
	@./bin/server

start: build run