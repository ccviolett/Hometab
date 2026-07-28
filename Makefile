BACKEND_DIR := apps/backend-go-fiber
FRONTEND_DIR := apps/frontend-react-vite

.PHONY: setup lint test test-e2e build check clean

setup:
	cd $(FRONTEND_DIR) && npm ci --no-audit
	cd $(BACKEND_DIR) && go mod download

lint:
	cd $(FRONTEND_DIR) && npm run lint
	@test -z "$$(gofmt -l $(BACKEND_DIR))"
	cd $(BACKEND_DIR) && go vet ./...

test:
	cd $(FRONTEND_DIR) && npm test
	cd $(BACKEND_DIR) && go test -race ./...

test-e2e:
	cd $(BACKEND_DIR) && go test -race ./e2e

build:
	$(MAKE) -C $(BACKEND_DIR) build

check: lint test
	cd $(FRONTEND_DIR) && npm run build
	cd $(BACKEND_DIR) && go build ./...

clean:
	$(MAKE) -C $(BACKEND_DIR) clean
	rm -rf $(FRONTEND_DIR)/dist
