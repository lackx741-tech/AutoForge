.PHONY: all build build-ckg build-saga build-cli run stop test docker up down clean install push

all: build

build: build-ckg build-saga build-cli
	@echo "Built: bin/ckg-service, bin/saga-supervisor, bin/autoforge"

# CKG requires CGO for SQLite
build-ckg:
	@mkdir -p bin
	cd cmd/ckg-service && go mod tidy && CGO_ENABLED=1 go build -o ../../bin/ckg-service .

build-saga:
	@mkdir -p bin
	cd cmd/saga-supervisor && go mod tidy && go build -o ../../bin/saga-supervisor .

build-cli:
	@mkdir -p bin
	cd cmd/cli && go mod tidy && go build -o ../../bin/autoforge .

run: build
	@echo "Starting services..."
	./bin/ckg-service &
	@sleep 1
	./bin/saga-supervisor &
	@echo "Services running on :8080 and :8081"
	@echo "Use: ./bin/autoforge submit <repo> <commit>"

stop:
	@pkill -f "bin/ckg-service" 2>/dev/null || true
	@pkill -f "bin/saga-supervisor" 2>/dev/null || true
	@echo "Services stopped"

test: build
	@echo "Testing services..."
	./bin/ckg-service &
	@sleep 2
	@curl -sf http://localhost:8080/health && echo " CKG healthy" || (echo " CKG failed" && exit 1)
	./bin/saga-supervisor &
	@sleep 1
	@curl -sf http://localhost:8081/health && echo " Saga healthy" || (echo " Saga failed" && exit 1)
	@pkill -f "bin/ckg-service" 2>/dev/null || true
	@pkill -f "bin/saga-supervisor" 2>/dev/null || true

docker:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

clean:
	rm -rf bin/
	docker compose down -v 2>/dev/null || true

install: build
	cp bin/autoforge /usr/local/bin/
	@echo "Installed autoforge CLI to /usr/local/bin"

push:
	git add .
	git commit -m "AutoForge update" || true
	git push origin main || echo "Push manually with: git push"
