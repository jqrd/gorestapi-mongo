EXECUTABLE := gorestapicmd
GITVERSION := $(shell git describe --dirty --always --tags --long)
GOPATH ?= ${HOME}/go
PACKAGENAME := $(shell cd src && go list -m -f '{{.Path}}')
TOOLS := ${GOPATH}/bin/mockery \
	${GOPATH}/bin/swag \
	${GOPATH}/bin/protoc-gen-go \
	${GOPATH}/bin/protoc-gen-gotag
SWAGGERSOURCE = $(wildcard src/gorestapi/*.go) \
	$(wildcard src/gorestapi/mainrpc/*.go)
GOSOURCE = $(shell find ./src -type f)


${EXECUTABLE}: tools proto ${GOSOURCE} swagger
	# Compiling...
	cd src && go build -ldflags "-X ${PACKAGENAME}/conf.Executable=${EXECUTABLE} -X ${PACKAGENAME}/conf.GitVersion=${GITVERSION}" -o ../${EXECUTABLE}


.PHONY: tools
tools: ${TOOLS}

${GOPATH}/bin/mockery:
	go install github.com/vektra/mockery/v3@v3.0.0-alpha.0

${GOPATH}/bin/swag:
	go install github.com/swaggo/swag/cmd/swag@latest

${GOPATH}/bin/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

${GOPATH}/bin/protoc-gen-gotag:
	go install github.com/srikrsna/protoc-gen-gotag


.PHONY: swagger
swagger: src/embed/public_html/api-docs/swagger.json

src/embed/public_html/api-docs/swagger.json: tools ${SWAGGERSOURCE}
	cd src && swag init --dir . --generalInfo gorestapi/swagger.go --exclude embed --output embed/public_html/api-docs --outputTypes json


.PHONY: mocks
mocks: tools src/mocks/MongoCollection.go src/mocks/DataStore.go

src/mocks/MongoCollection.go: src/store/mongodb/collection.go
	cd src && mockery --dir ./store/mongodb --name MongoCollection

src/mocks/DataStore.go: src/gorestapi/datastore.go
	cd src && mockery --dir ./gorestapi --name DataStore


.PHONY: proto
proto: src/model/db/db.pb.go \
		src/model/svc/svc.pb.go

.PHONY: proto-clean
proto-clean:
	@rm -f src/model/db/db.pb.go
	@rm -f src/model/svc/svc.pb.go
	@rm -f src/model/common.pb.go
	@rm -f src/model/tagger/tagger.pb.go

src/model/db/db.pb.go: src/model/tagger/tagger.pb.go \
		src/model/common.pb.go \
		src/model/db/db.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/db/db.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=auto="bson-as-camel+json-as-camel":. model/db/db.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=xxx="bson+\"-\" json+\"-\"":. model/db/db.proto

src/model/svc/svc.pb.go: src/model/tagger/tagger.pb.go \
		src/model/common.pb.go \
		src/model/svc/svc.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/svc/svc.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=auto="bson-as-camel+json-as-camel":. model/svc/svc.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=xxx="bson+\"-\" json+\"-\"":. model/svc/svc.proto

src/model/common.pb.go: src/model/common.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/common.proto
# TODO without the full package path (e.g. with relative path), wrong import path is generated in the files that import this, but with it the file gets placed in an unexpected place
	@mv src/github.com/jqrd/gorestapi-mongo/model/common.pb.go src/model/common.pb.go
	@rm -r src/github.com

src/model/tagger/tagger.pb.go: src/model/tagger/tagger.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/tagger/tagger.proto
# TODO without the full package path (e.g. with relative path), wrong import path is generated in the files that import this, but with it the file gets placed in an unexpected place
	@mv src/github.com/jqrd/gorestapi-mongo/model/tagger/tagger.pb.go src/model/tagger/tagger.pb.go
	@rm -r src/github.com


.PHONY: test
test: tools mocks
	cd src && go test -cover ./...

.PHONY: deps
deps:
	# Fetching dependancies...
	cd src && go get -d -v # Adding -u here will break CI

.PHONY: lint
lint:
	docker run --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v1.27.0 golangci-lint run -v --timeout 5m

.PHONY: hadolint
hadolint:
	docker run -it --rm -v ${PWD}/Dockerfile:/Dockerfile hadolint/hadolint:latest hadolint --ignore DL3018 Dockerfile

.PHONY: relocate
relocate:
	@test ${TARGET} || ( echo ">> TARGET is not set. Use: make relocate TARGET=<target>"; exit 1 )
	$(eval ESCAPED_PACKAGENAME := $(shell echo "${PACKAGENAME}" | sed -e 's/[\/&]/\\&/g'))
	$(eval ESCAPED_TARGET := $(shell echo "${TARGET}" | sed -e 's/[\/&]/\\&/g'))
	# Renaming package ${PACKAGENAME} to ${TARGET}
	@grep -rlI '${PACKAGENAME}' * | xargs -i@ sed -i 's/${ESCAPED_PACKAGENAME}/${ESCAPED_TARGET}/g' @
	# Complete... 
	# NOTE: This does not update the git config nor will it update any imports of the root directory of this project.
