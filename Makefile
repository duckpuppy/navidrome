GO_VERSION=1.13
NODE_VERSION=v13.7.0

GIT_SHA=$(shell git rev-parse --short HEAD)

.PHONY: dev
dev: check_env
	@goreman -f Procfile.dev -b 4533 start

.PHONY: server
server: check_go_env
	@reflex -d none -c reflex.conf

.PHONY: watch
watch: check_go_env
	ginkgo watch -notify ./...

.PHONY: test
test: check_go_env
	go test ./... -v
#	@(cd ./ui && npm test -- --watchAll=false)

.PHONY: testall
testall: check_go_env test
	@(cd ./ui && npm test -- --watchAll=false)

.PHONY: setup
setup: Jamstash-master
	@which goconvey   || (echo "Installing GoConvey" && GO111MODULE=off go get -u github.com/smartystreets/goconvey)
	@which wire       || (echo "Installing Wire"     && GO111MODULE=off go get -u github.com/google/wire/cmd/wire)
	@which go-bindata || (echo "Installing BinData"  && GO111MODULE=off go get -u github.com/go-bindata/go-bindata/...)
	@which reflex     || (echo "Installing Reflex"   && GO111MODULE=off go get -u github.com/cespare/reflex)
	@which goreman    || (echo "Installing Goreman"  && GO111MODULE=off go get -u github.com/mattn/goreman)
	@which ginkgo     || (echo "Installing Ginkgo"   && GO111MODULE=off go get -u github.com/onsi/ginkgo/ginkgo)
	@which goose      || (echo "Installing Goose"    && GO111MODULE=off go get -u github.com/pressly/goose/cmd/goose)
	go mod download
	@(cd ./ui && npm ci)

.PHONY: static
static:
	cd static && go-bindata -fs -prefix "static" -nocompress -ignore="\\\*.go" -pkg static .

Jamstash-master:
	wget -N https://github.com/tsquillario/Jamstash/archive/master.zip
	unzip -o master.zip
	rm master.zip

.PHONE: check_env
check_env: check_go_env check_node_env

.PHONY: check_go_env
check_go_env:
	@(hash go) || (echo "\nERROR: GO environment not setup properly!\n"; exit 1)
	@go version | grep -q $(GO_VERSION) || (echo "\nERROR: Please upgrade your GO version\n"; exit 1)

.PHONY: check_node_env
check_node_env:
	@(hash node) || (echo "\nERROR: Node environment not setup properly!\n"; exit 1)
	@node --version | grep -q $(NODE_VERSION) || (echo "\nERROR: Please check your Node version. Should be $(NODE_VERSION)\n"; exit 1)

.PHONY: build
build: check_go_env
	go build -ldflags="-X github.com/deluan/navidrome/consts.gitSha=$(GIT_SHA) -X github.com/deluan/navidrome/consts.gitTag=master"

.PHONY: buildall
buildall: check_env
	@(cd ./ui && npm run build)
	go-bindata -fs -prefix "ui/build" -tags embed -nocompress -pkg assets -o assets/embedded_gen.go ui/build/...
	go build -ldflags="-X github.com/deluan/navidrome/consts.gitSha=$(GIT_SHA) -X github.com/deluan/navidrome/consts.gitTag=master" -tags=embed

.PHONY: release
release:
	@if [[ ! "${V}" =~ ^[0-9]+\.[0-9]+\.[0-9]+.*$$ ]]; then echo "Usage: make release V=X.X.X"; exit 1; fi
	go mod tidy
	@if [ -n "`git status -s`" ]; then echo "\n\nThere are pending changes. Please commit or stash first"; exit 1; fi
	make test
	git tag v${V}
	git push origin v${V}

.PHONY: dist
dist:
	 docker run -it -v $(PWD):/github/workspace -w /github/workspace bepsays/ci-goreleaser:1.13-4 goreleaser release --rm-dist --skip-publish --snapshot
