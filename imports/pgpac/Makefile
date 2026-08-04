APP := pgpac
CMD := ./cmd/pgpac
SAMPLE_PROJECT := testdata/sample/sample.pgpac
OUT_DIR := out
SAMPLE_PACKAGE := $(OUT_DIR)/SampleProject.pgpkg
RELEASE_DIR := .out
VERSION ?= dev
version ?= $(VERSION)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
commit ?= $(COMMIT)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
build_date ?= $(BUILD_DATE)
LDFLAGS := -X 'main.version=$(version)' -X 'main.commit=$(commit)' -X 'main.buildDate=$(build_date)'
NATIVE_GOOS := $(shell go env GOOS)

.PHONY: build test sample package clean verify-changelog release-prepare release-build publish publish-darwin publish-linux publish-all clean-release website website-install website-build website-version website-typecheck

build:
	go build -ldflags "$(LDFLAGS)" -o $(APP) $(CMD)

test:
	go test ./...

sample: $(SAMPLE_PACKAGE)

package: $(SAMPLE_PACKAGE)

$(SAMPLE_PACKAGE):
	go run -ldflags "$(LDFLAGS)" $(CMD) build --project $(SAMPLE_PROJECT) --output $(OUT_DIR)/

verify-changelog:
	REQUIRE_CHANGELOG_ALWAYS=true ENFORCE_UNRELEASED_BULLET=true .github/scripts/check-unreleased-changelog.sh

release-prepare:
	./.github/scripts/release.sh "$(version)" "$(dryrun)"

release-build: clean-release publish-all

publish: clean-release
	mkdir -p $(RELEASE_DIR)/current
	go build -ldflags "$(LDFLAGS)" -o $(RELEASE_DIR)/current/$(APP) $(CMD)

publish-darwin:
	@if [ "$(NATIVE_GOOS)" != "darwin" ]; then echo "publish-darwin must run on a native darwin runner"; exit 1; fi
	COMMIT="$(commit)" BUILD_DATE="$(build_date)" ./.github/scripts/build-release-archive.sh "$(version)" "$(RELEASE_DIR)"

publish-linux:
	@if [ "$(NATIVE_GOOS)" != "linux" ]; then echo "publish-linux must run on a native linux runner"; exit 1; fi
	COMMIT="$(commit)" BUILD_DATE="$(build_date)" ./.github/scripts/build-release-archive.sh "$(version)" "$(RELEASE_DIR)"

publish-all: test publish-$(NATIVE_GOOS)

clean-release:
	rm -rf $(RELEASE_DIR)

website-install:
	cd website && npm ci

website-build:
	cd website && npm run build

website:
	cd website && npm run start

website-version:
	cd website && npm run version-docs -- $(version)

website-typecheck:
	cd website && npm run typecheck

clean: clean-release
	rm -rf $(OUT_DIR) $(APP)
