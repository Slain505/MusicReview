DB_DSN ?= postgres://app:app@localhost:5432/musicreview?sslmode=disable

.PHONY: run migrate-up migrate-down migrate-version

run:
	DB_DSN='$(DB_DSN)' go run ./cmd/api

migrate-up:
	migrate -path migrations -database '$(DB_DSN)' up

migrate-down:
	migrate -path migrations -database '$(DB_DSN)' down 1

migrate-version:
	migrate -path migrations -database '$(DB_DSN)' version