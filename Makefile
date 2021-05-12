migrate = migrate -path db/migrations 
migrate += -database "postgres://postgres:secret@localhost:49174/postgres?sslmode=disable"

compose = docker-compose

migrate-up:
	$(migrate) -verbose up

migrate-down:
	$(migrate) -verbose down

build: 
	go build -o myriadcode cmd/web/main.go

run: build
	./myriadcode

docker:
	$(compose) up

docker-build:
	$(compose) build

docker-dev:
	$(compose) -f docker-compose.dev.yml up --build