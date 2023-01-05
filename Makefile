SHELL := /bin/bash

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



DOCKER = $(shell [[ $(shell docker ps 2>&1 | grep "permission denied" | wc -c | sed 's/ //g' ) -eq 0 ]] && echo "docker" || echo "sudo docker" )
MD5 = $(shell [[ $(shell echo "1" | md5sum 2>&1 | grep "command not found" | wc -c | sed 's/ //g' ) -eq 0 ]] && echo "md5sum" || echo "md5" )


.PHONY: build
build: ${EXECUTABLE}


.PHONY: which
which:
	@echo "DOCKER =  ${DOCKER}"
	@echo "MD5 =     ${MD5}"


.PHONY: build-docker
build-docker:
	# TODO make gorestapi docker image, deploy dev infra, deploy gorestapi container

.PHONY: infra-dev
infra-dev: infra-dev-up


${EXECUTABLE}: tools proto mocks ${GOSOURCE} swagger
	# Compiling...
	cd src && go build -ldflags "-X ${PACKAGENAME}/conf.Executable=${EXECUTABLE} -X ${PACKAGENAME}/conf.GitVersion=${GITVERSION}" -o ../${EXECUTABLE}


.PHONY: run
run: infra/dev/stack.yml
	$(eval MONGO_U = $(shell grep -oP "(?<=MONGO_INITDB_ROOT_USERNAME: )(.*)" infra/dev/stack.yml))
	$(eval MONGO_P = $(shell grep -oP "(?<=MONGO_INITDB_ROOT_PASSWORD: )(.*)" infra/dev/stack.yml))
	Mongo.Username=${MONGO_U} Mongo.Password=${MONGO_P} ${EXECUTABLE} api


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
	# Fetching dependencies...
	cd src && go get -d -v # Adding -u here will break CI

.PHONY: lint
lint:
	docker run --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v1.27.0 golangci-lint run -v --timeout 5m

.PHONY: hadolint
hadolint:
	docker run -it --rm -v ${PWD}/Dockerfile:/Dockerfile hadolint/hadolint:latest hadolint --ignore DL3018 Dockerfile


.PHONY: infra-dev-up
infra-dev-up: infra/dev/docker-compose.yml infra/dev/.env-update
	[ -d infra/dev/data ] || ( mkdir infra/dev/data && chmod 777 infra/dev/data )
	cd infra/dev && ${DOCKER} compose up --detach --wait

infra/dev/.env-update: infra/dev/.env
	# Env file was created/modified, ensuring containers get recreated...
	cd infra/dev && ${DOCKER} compose rm -s -f -v
	@touch infra/dev/.env-update

infra/dev/.env:
	@touch infra/dev/.env
	$(shell echo MONGO_HOST=mongo >> infra/dev/.env)
	$(shell echo MONGO_PORT=27017 >> infra/dev/.env)
	$(shell echo MONGO_USR=root >> infra/dev/.env)
	$(shell echo MONGO_PWD=$(shell echo mongo $RANDOM | ${MD5} | head -c 20; echo) >> infra/dev/.env)
	$(shell echo MONGO_EXPRESS_HOST=mongo_express >> infra/dev/.env)
	$(shell echo MONGO_EXPRESS_PORT=8081 >> infra/dev/.env)
	$(shell echo MONGO_EXPRESS_USR=admin >> infra/dev/.env)
	$(shell echo MONGO_EXPRESS_PWD=$(shell echo express $RANDOM | ${MD5} | head -c 20; echo) >> infra/dev/.env)
	# Passwords & other settings generated into env file infra/dev/.env

.PHONY: infra-dev-clean
infra-dev-clean:
	cd infra/dev && ${DOCKER} compose rm -s -f -v
	sudo rm -rf infra/dev/data


.PHONY: infra-stage
infra-stage:
	#TODO with terraform probably


.PHONY: relocate
relocate:
	@test ${TARGET} || ( echo ">> TARGET is not set. Use: make relocate TARGET=<target>"; exit 1 )
	$(eval ESCAPED_PACKAGENAME := $(shell echo "${PACKAGENAME}" | sed -e 's/[\/&]/\\&/g'))
	$(eval ESCAPED_TARGET := $(shell echo "${TARGET}" | sed -e 's/[\/&]/\\&/g'))
	# Renaming package ${PACKAGENAME} to ${TARGET}
	@grep -rlI '${PACKAGENAME}' * | xargs -i@ sed -i 's/${ESCAPED_PACKAGENAME}/${ESCAPED_TARGET}/g' @
	# Complete... 
	# NOTE: This does not update the git config nor will it update any imports of the root directory of this project.
