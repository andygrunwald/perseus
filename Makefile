CLONE_URL = github.com/andygrunwald/perseus
VERSION = `git rev-parse --abbrev-ref HEAD 2>/dev/null`
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`
LDFLAGS = -ldflags "-X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildDate=${BUILD_DATE}"

build:
	go build ${LDFLAGS} ${CLONE_URL}/cmd/perseus

install:
	go install ${LDFLAGS} ${CLONE_URL}/cmd/perseus

test:
	GOMAXPROCS=4 GORACE="halt_on_error=1" go test -race -v ./...