up:
	docker-compose up --build

test:
	go test ./internal/service/ -v
