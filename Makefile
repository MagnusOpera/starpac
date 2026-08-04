TOOLS := pgpac d1pac
VERSION ?= dev
version ?= $(VERSION)
tool ?=
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
commit ?= $(COMMIT)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
build_date ?= $(BUILD_DATE)
RELEASE_DIR := .out

.PHONY: build test pgpac-build d1pac-build sample clean verify-changelog release-build release-prepare release-smoke release-test website website-install website-typecheck website-build website-version

build: pgpac-build d1pac-build

test:
	go test ./...
	./.github/scripts/test-release.sh

pgpac-build:
	go build -ldflags "-X main.version=$(version) -X main.commit=$(commit) -X main.buildDate=$(build_date)" -o pgpac ./cmd/pgpac

d1pac-build:
	go build -ldflags "-X main.version=$(version) -X main.commit=$(commit) -X main.buildDate=$(build_date)" -o d1pac ./cmd/d1pac

sample:
	mkdir -p out
	go run ./cmd/pgpac build --project products/pgpac/testdata/sample/sample.pgpac --output out/
	go run ./cmd/d1pac build --project products/d1pac/testdata/sample/sample.d1pac --output out/

verify-changelog:
	./.github/scripts/check-unreleased-changelog.sh

release-build:
	@case "$(tool)" in pgpac|d1pac) ;; *) echo "Usage: make release-build tool=<pgpac|d1pac> version=X.Y.Z"; exit 2;; esac
	./.github/scripts/build-release-archive.sh "$(tool)" "$(version)" "$(RELEASE_DIR)"

release-prepare:
	./.github/scripts/release.sh "$(version)" "$(dryrun)"

release-smoke:
	./.github/scripts/release-smoke.sh

release-test:
	./.github/scripts/test-release.sh

website-install:
	rm -rf website/node_modules/.cache
	cd website && npm ci

website-typecheck:
	cd website && npm run typecheck

website-build:
	cd website && npm run build

website-version:
	cd website && npm run version-docs -- "$(version)"

website:
	cd website && npm run start

clean:
	rm -rf .out out website/build pgpac d1pac
