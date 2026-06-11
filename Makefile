.PHONY: build-datetime-app test frontend-build frontend-lint

build-datetime-app:
	docker build -t matryoshka-datetime-app:latest examples/datetime-go

test:
	go test ./internal/graphql/... ./internal/manager/...

frontend-build:
	cd frontend && bun run build

frontend-lint:
	cd frontend && bun run lint
