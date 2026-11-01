# One-command workflows. Requires Go ≥1.27 (auto-downloaded via toolchain), Node ≥22, PostgreSQL 16.
SHELL := /bin/bash
HARNESS_ROOT := $(shell git rev-parse --show-toplevel)
DB_URL ?= postgres://postgres@127.0.0.1:5432/realmforge?sslmode=disable
WORLDS ?= ember-1:ashen-crowns:Ember Vale (x50):50,void-1:void-frontier:Void Frontier Demo:10

.PHONY: verify-backup db dev server client build test test-go test-client e2e packs lint backup restore docker sim
.PHONY: harness-context harness-doctor harness-test harness-check check-spec

harness-context: ## show shared agent context
	cd "$(HARNESS_ROOT)" && python3 scripts/harness/harness.py context

harness-doctor: ## check local agent dependencies and config without installing anything
	cd "$(HARNESS_ROOT)" && python3 scripts/harness/harness.py doctor

harness-test: ## test both agent adapters without accounts or PostgreSQL
	cd "$(HARNESS_ROOT)" && PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/harness -p 'test_*.py'

check-spec: ## validate My Lands specification, cards and coverage
	cd "$(HARNESS_ROOT)" && python3 scripts/check_mylands_spec.py

harness-check: ## shared completion gate (E2E remains required for user-facing flows)
	$(MAKE) --no-print-directory -C "$(HARNESS_ROOT)" harness-test
	$(MAKE) --no-print-directory -C "$(HARNESS_ROOT)" lint
	$(MAKE) --no-print-directory -C "$(HARNESS_ROOT)" test
	$(MAKE) --no-print-directory -C "$(HARNESS_ROOT)" packs

db:            ## start native PostgreSQL (when Docker is unavailable)
	./scripts/dev-postgres.sh start

server:        ## run API + worker against $(DB_URL)
	cd server && RF_DATABASE_URL="$(DB_URL)" RF_PACKS_DIR=../packs RF_CLIENT_DIR=../client/dist RF_WORLDS="$(WORLDS)" go run ./cmd/server

client:        ## Vite dev server with API proxy on :5173
	cd client && npm run dev

dev: db build server   ## start db, build client, run server on :8080

build:         ## production builds
	cd client && npm ci --no-audit --no-fund && npm run build
	cd server && go build -o bin/server ./cmd/server && go build -o bin/packtool ./cmd/packtool

packs:         ## validate all game packs
	cd server && go run ./cmd/packtool summary ../packs/*/

sim:           ## balance simulation of the first day (no database needed)
	cd server && go run ./cmd/sim -pack ../packs/ashen-crowns,../packs/void-frontier -out ../docs/reports/BALANCE_SIM_TABLES.md

lint:
	cd server && go vet ./... && test -z "$$(gofmt -l .)"
	cd client && npm run typecheck

test: test-go test-client  ## all automated tests (needs PostgreSQL)

test-go:
	cd server && RF_REQUIRE_DB=1 go test -race -count=1 ./...

test-client:
	cd client && npm test

e2e:           ## browser E2E against a running server (make dev in another shell)
	cd client && node e2e/play.mjs http://127.0.0.1:8080 $${SHOTS:-/tmp}

backup:        ## logical backup to backups/<timestamp>.dump
	./scripts/backup.sh "$(DB_URL)"

restore:       ## restore FILE=backups/x.dump into $(DB_URL)
	./scripts/restore.sh "$(DB_URL)" "$(FILE)"

docker:
	docker compose up --build

verify-backup: ## backup + restore into scratch DB + compare counts
	./scripts/verify-backup.sh "$(DB_URL)"
