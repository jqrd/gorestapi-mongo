EXECUTABLE := gorestapicmd
GITVERSION := $(shell git describe --dirty --always --tags --long)
GOPATH ?= ${HOME}/go
PACKAGENAME := $(shell go list -m -f '{{.Path}}')
TOOLS := ${GOPATH}/bin/mockery \
	${GOPATH}/bin/swag \
	${GOPATH}/bin/protoc-gen-go \
	${GOPATH}/bin/protoc-gen-gotag
SWAGGERSOURCE = $(wildcard gorestapi/*.go) \
	$(wildcard gorestapi/mainrpc/*.go)
GOSOURCE = go.mod \
	go.sum \
	$(shell find . -type f -name '*.go') \
	$(shell find ./embed -type f)


${EXECUTABLE}: tools ${GOSOURCE} swagger
	# Compiling...
	go build -ldflags "-X ${PACKAGENAME}/conf.Executable=${EXECUTABLE} -X ${PACKAGENAME}/conf.GitVersion=${GITVERSION}" -o ${EXECUTABLE}


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
swagger: embed/public_html/api-docs/swagger.json

embed/public_html/api-docs/swagger.json: tools ${SWAGGERSOURCE}
	swag init --dir . --generalInfo gorestapi/swagger.go --exclude embed --output embed/public_html/api-docs --outputTypes json


.PHONY: mocks
mocks: tools mocks/MongoCollection.go mocks/DataStore.go

mocks/MongoCollection.go: store/mongodb/collection.go
	mockery --dir ./store/mongodb --name MongoCollection

mocks/DataStore.go: gorestapi/datastore.go
	mockery --dir ./gorestapi --name DataStore


.PHONY: proto
proto: model/db/db.pb.go \
		model/svc/svc.pb.go

.PHONY: proto-clean
proto-clean:
	@rm -f model/db/db.pb.go
	@rm -f model/svc/svc.pb.go
	@rm -f model/common.pb.go
	@rm -f model/tagger/tagger.pb.go

model/db/db.pb.go: model/tagger/tagger.pb.go \
		model/common.pb.go \
		model/db/db.proto
	protoc -I /usr/local/include -I . --go_out=:. model/db/db.proto
	protoc -I /usr/local/include -I . --gotag_out=auto="bson-as-camel+json-as-camel":. model/db/db.proto
	protoc -I /usr/local/include -I . --gotag_out=xxx="bson+\"-\" json+\"-\"":. model/db/db.proto

model/svc/svc.pb.go: model/tagger/tagger.pb.go \
		model/common.pb.go \
		model/svc/svc.proto
	protoc -I /usr/local/include -I . --go_out=:. model/svc/svc.proto
	protoc -I /usr/local/include -I . --gotag_out=auto="bson-as-camel+json-as-camel":. model/svc/svc.proto
	protoc -I /usr/local/include -I . --gotag_out=xxx="bson+\"-\" json+\"-\"":. model/svc/svc.proto

model/common.pb.go: model/common.proto
	protoc -I /usr/local/include -I . --go_out=:. model/common.proto
# TODO without the full package path (e.g. with relative path), wrong import path is generated in the files that import this, but with it the file gets placed in an unexpected place
	@mv github.com/jqrd/gorestapi-mongo/model/common.pb.go model/common.pb.go
	@rm -r github.com

model/tagger/tagger.pb.go: model/tagger/tagger.proto
	protoc -I /usr/local/include -I . --go_out=:. model/tagger/tagger.proto
# TODO without the full package path (e.g. with relative path), wrong import path is generated in the files that import this, but with it the file gets placed in an unexpected place
	@mv github.com/jqrd/gorestapi-mongo/model/tagger/tagger.pb.go model/tagger/tagger.pb.go
	@rm -r github.com


.PHONY: test
test: tools mocks
	go test -cover ./...

.PHONY: deps
deps:
	# Fetching dependancies...
	go get -d -v # Adding -u here will break CI

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
