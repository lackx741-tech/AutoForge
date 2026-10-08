.PHONY: all build run test clean docker push

all: build

build:
	@echo "Building AutoForge services..."
	@mkdir -p bin
	cd cmd/ckg-service && go mod tidy && go build -o ../../bin/ckg-service
	cd cmd/saga-supervisor && go mod tidy && go build -o ../../bin/saga-supervisor
	cd cmd/cli && go mod tidy && go build -o ../../bin/autoforge
	@echo "Built: bin/ckg-service, bin/saga-supervisor, bin/autoforge"

run: build
	@echo "Starting services..."
	./bin/ckg-service &
	./bin/saga-supervisor &
	@echo "Services running on :8080 and :8081"
	@echo "Use: ./bin/autoforge submit <repo> <commit>"

docker:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

test:
	curl -f http://localhost:8080/health
	curl -f http://localhost:8081/health

clean:
	rm -rf bin/
	docker-compose down -v || true

install: build
	cp bin/autoforge /usr/local/bin/
	@echo "Installed autoforge CLI to /usr/local/bin"

push:
	git add .
	git commit -m "AutoForge initial commit" || true
	git push origin main || echo "Push manually with: git push"
