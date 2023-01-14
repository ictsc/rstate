
build:
	go build -ldflags '-s -w' -o ./bin/server cmd/terraform-server/*.go
up:
	docker-compose up -d --build
log:
	docker-compose logs --tail=50
