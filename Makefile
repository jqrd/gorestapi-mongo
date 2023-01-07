SHELL := /bin/bash
GOPATH ?= ${HOME}/go


EXECUTABLE := gorestapicmd
GIT_VERSION := $(shell git describe --dirty --always --tags --long)
PACKAGE_NAME := $(shell cd src && go list -m -f '{{.Path}}')
CONTAINER_NAME := $(shell echo "${PACKAGE_NAME}" | grep -Eo '(^|/)[^/]+$$' | sed s!/!!)

# go tools:
MOCKERY := ${GOPATH}/bin/mockery
SWAG :=	${GOPATH}/bin/swag
PROTOC_GEN_GO := ${GOPATH}/bin/protoc-gen-go
PROTOC_GEN_GOTAG := ${GOPATH}/bin/protoc-gen-gotag

# external tools:
PROTOC := $(if $(shell which protoc),$(shell which protoc),"protoc_not_found")
PROTOC_MIN_VERSION := 3.18.1

# env-dependent tools:
DOCKER := $(shell [[ $(shell docker ps 2>&1 | grep "permission denied" | wc -c | sed 's/ //g' ) -eq 0 ]] && echo "docker" || echo "sudo docker" )
MD5 := $(shell [[ $(shell echo "1" | md5sum 2>&1 | grep "command not found" | wc -c | sed 's/ //g' ) -eq 0 ]] && echo "md5sum" || echo "md5" )
SED := $(shell [[ $(shell sed --help 2>&1 | grep -F -- "-i[SUFFIX]" | wc -c | sed 's/ //g' ) -eq 0 ]] && echo "sed -I ''" || echo "sed -i" )


.PHONY: build
build: proto mocks swagger ${EXECUTABLE}

.PHONY: dev-infra
dev-infra: dev-infra-up

.PHONY: run-docker
run-docker: dev-docker-run

.PHONY: clean
clean: dev-docker-stop dev-infra-clean proto-clean mocks-clean swagger-clean ${EXECUTABLE}-clean


.PHONY: which
which:
	@echo "MD5              =  ${MD5}"
	@echo "SED              =  ${SED}"
	@echo "MOCKERY          =  ${MOCKERY}"
	@echo "SWAG             =  ${SWAG}"
	@echo "PROTOC           =  ${PROTOC}"
	@echo "PROTOC_GEN_GO    =  ${PROTOC_GEN_GO}"
	@echo "PROTOC_GEN_GOTAG =  ${PROTOC_GEN_GOTAG}"
	@echo "DOCKER           =  ${DOCKER}"


.PHONY: run
run: ${EXECUTABLE} dev-infra-up
	$(eval $(shell grep -Eoh "MONGO_PORT=.*" infra/dev/.env))
	$(eval $(shell grep -Eoh "MONGO_USR=.*" infra/dev/.env))
	$(eval $(shell grep -Eoh "MONGO_PWD=.*" infra/dev/.env))
	DATABASE_HOST=localhost \
		DATABASE_PORT=${MONGO_PORT} \
		DATABASE_USERNAME=${MONGO_USR} \
		DATABASE_PASSWORD=${MONGO_PWD} \
		./${EXECUTABLE} api

# Tools install/check
${MOCKERY}:
	cd src && go install github.com/vektra/mockery/v3@v3.0.0-alpha.0
${SWAG}:
	cd src && go install github.com/swaggo/swag/cmd/swag@latest
${PROTOC_GEN_GO}:
	cd src && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
${PROTOC_GEN_GOTAG}:
	cd src && go install github.com/srikrsna/protoc-gen-gotag
${PROTOC}:
	@echo "Protoc not found, please install protoc." && exit 1


SWAGGER_OUT := src/embed/public_html/api-docs/swagger.json

.PHONY: swagger
swagger: ${SWAGGER_OUT}

.PHONY: swagger-clean
swagger-clean:
	rm -f src/embed/public_html/api-docs/swagger.json

src/embed/public_html/api-docs/swagger.json: .EXTRA_PREREQS = ${SWAG}
src/embed/public_html/api-docs/swagger.json: $(wildcard src/gorestapi/*.go) \
		$(wildcard src/gorestapi/mainrpc/*.go)
	cd src && swag init --dir . \
		--generalInfo gorestapi/swagger.go \
		--exclude embed \
		--output embed/public_html/api-docs \
		--outputTypes json


MOCKS_OUT := src/mocks/MongoCollection.go \
	src/mocks/DataStore.go

.PHONY: mocks
mocks: ${MOCKS_OUT}

.PHONY: mocks-clean
mocks-clean:
	rm -f src/mocks/MongoCollection.go src/mocks/DataStore.go

src/mocks/MongoCollection.go: .EXTRA_PREREQS = ${MOCKERY}
src/mocks/MongoCollection.go: src/store/mongodb/collection.go
	cd src && mockery --dir ./store/mongodb --name MongoCollection

src/mocks/DataStore.go: .EXTRA_PREREQS = ${MOCKERY}
src/mocks/DataStore.go: src/gorestapi/datastore.go
	cd src && mockery --dir ./gorestapi --name DataStore


PROTOC_CHECK_VERSION := infra/dev/.protoc.version
PROTO_OUT := src/model/db/db.pb.go \
		src/model/svc/svc.pb.go

.PHONY: proto
proto: ${PROTOC_CHECK_VERSION} ${PROTO_OUT}

${PROTOC_CHECK_VERSION}: ${PROTOC}
	$(eval PROTOC_VERSION=$(shell protoc --version | sed 's/.* //'))
	@(echo "${PROTOC_MIN_VERSION}" && echo "${PROTOC_VERSION}") | (sort -V -c > /dev/null 2>&1 && echo "You have protoc ${PROTOC_VERSION} >= min version ${PROTOC_MIN_VERSION} ✓") || (echo "⛔️ minimum required version of protoc is ${PROTOC_MIN_VERSION}, please upgrade (you have ${PROTOC_VERSION})"; echo "(older ones reference deprecated protobuf package from github.com instead of google.golang.org)" && exit 1)
	@echo "${PROTOC_VERSION}" > ${PROTOC_CHECK_VERSION}

.PHONY: proto-clean
proto-clean:
	@rm -f src/model/db/db.pb.go
	@rm -f src/model/svc/svc.pb.go
	@rm -f src/model/common.pb.go
	@rm -f src/model/tagger/tagger.pb.go

src/model/db/db.pb.go: .EXTRA_PREREQS = ${PROTOC_CHECK_VERSION}
src/model/db/db.pb.go: src/model/tagger/tagger.pb.go \
		src/model/common.pb.go \
		src/model/db/db.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/db/db.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=auto="bson-as-camel+json-as-camel":. model/db/db.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=xxx="bson+\"-\" json+\"-\"":. model/db/db.proto

src/model/svc/svc.pb.go: .EXTRA_PREREQS = ${PROTOC_CHECK_VERSION}
src/model/svc/svc.pb.go: src/model/tagger/tagger.pb.go \
		src/model/common.pb.go \
		src/model/svc/svc.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/svc/svc.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=auto="bson-as-camel+json-as-camel":. model/svc/svc.proto
	cd src && protoc -I /usr/local/include -I . --gotag_out=xxx="bson+\"-\" json+\"-\"":. model/svc/svc.proto

src/model/common.pb.go: .EXTRA_PREREQS = ${PROTOC_CHECK_VERSION}
src/model/common.pb.go: src/model/common.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/common.proto
# TODO without the full package path (e.g. with relative path), wrong import path is generated in the files that import this, but with it the file gets placed in an unexpected place
	@mv src/${PACKAGE_NAME}/model/common.pb.go src/model/common.pb.go
	@rm -r src/$(shell echo "${PACKAGE_NAME}" | grep -Eoh "^[^/$$]+")

src/model/tagger/tagger.pb.go: .EXTRA_PREREQS = ${PROTOC_CHECK_VERSION}
src/model/tagger/tagger.pb.go: src/model/tagger/tagger.proto
	cd src && protoc -I /usr/local/include -I . --go_out=:. model/tagger/tagger.proto
# TODO without the full package path (e.g. with relative path), wrong import path is generated in the files that import this, but with it the file gets placed in an unexpected place
	@mv src/${PACKAGE_NAME}/model/tagger/tagger.pb.go src/model/tagger/tagger.pb.go
	@rm -r src/$(shell echo "${PACKAGE_NAME}" | grep -Eoh "^[^/$$]+")


${EXECUTABLE}: $(shell find ./src -type f) \
		${PROTO_OUT} \
		${MOCKS_OUT} \
		${SWAGGER_OUT}
	# Compiling...
	echo proto deps = ${PROTO_OUT}
	echo mocks deps = ${MOCKS_OUT}
	echo swagger deps = ${SWAGGER_OUT}
	cd src && go build \
		-ldflags "-X ${PACKAGE_NAME}/conf.Executable=${EXECUTABLE} \
		          -X ${PACKAGE_NAME}/conf.GitVersion=${GIT_VERSION}" \
		-o ../${EXECUTABLE}

PHONY: ${EXECUTABLE}-clean
${EXECUTABLE}-clean:
	rm -f ${EXECUTABLE}


.PHONY: test
test: ${EXECUTABLE}
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


.PHONY: dev-infra-up
dev-infra-up: infra/dev/docker-compose.yml infra/dev/.env-update
	[ -d infra/dev/data ] || ( mkdir infra/dev/data && chmod 777 infra/dev/data )
	cd infra/dev && ${DOCKER} compose up --detach --wait

infra/dev/.env-update: infra/dev/.env
	# Env file was created/modified, ensuring containers get recreated...
	cd infra/dev && ${DOCKER} compose rm -s -f -v
#	${DOCKER} secret rm MONGO_USR MONGO_PWD || true
#	# Creating docker secrets: MONGO_USR MONGO_PWD
#	$(shell cat infra/dev/.env | grep "MONGO_USR" | sed s/MONGO_USR=// | ${DOCKER} secret create MONGO_USR - > /dev/null)
#	$(shell cat infra/dev/.env | grep "MONGO_PWD" | sed s/MONGO_PWD=// | ${DOCKER} secret create MONGO_PWD -> /dev/null)
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

.PHONY: dev-infra-clean
dev-infra-clean:
	cd infra/dev && ${DOCKER} compose rm -s -f -v
#	${DOCKER} secret rm MONGO_USR MONGO_PWD || true
	[ -d infra/dev/data ] && sudo rm -rf infra/dev/data || true
	# `make dev-infra-clean` not cleaning infra/dev/.env and infra/dev/.env-update to keep generated passwords.
	# Use `make dev-infra-clean-env` to clean them.

.PHONY: dev-infra-clean-env
dev-infra-clean-env:
	rm -f infra/dev/.env infra/dev/.env-update

.PHONY: dev-docker-build
dev-docker-build: dev-infra-clean
	${DOCKER} build . -f infra/dev/Dockerfile \
		--build-arg EXECUTABLE=${EXECUTABLE} \
		--build-arg GIT_VERSION=${GIT_VERSION} \
		-t ${CONTAINER_NAME}:dev

.PHONY: dev-docker-run
dev-docker-run: dev-docker-stop \
		dev-docker-build \
		dev-infra-up \
		dev-docker-start

.PHONY: dev-docker-start
dev-docker-start:
# TODO convert to swarm and pass/read these as secrets
	$(eval $(shell grep -Eoh "MONGO_HOST=.*" infra/dev/.env))
	$(eval $(shell grep -Eoh "MONGO_PORT=.*" infra/dev/.env))
	$(eval $(shell grep -Eoh "MONGO_USR=.*" infra/dev/.env))
	$(eval $(shell grep -Eoh "MONGO_PWD=.*" infra/dev/.env))
	$(eval PORT=8080)
	${DOCKER} run -d \
		-e server.port=${PORT} \
		-p ${PORT}:${PORT}/tcp \
		-e DATABASE_HOST=${MONGO_HOST} \
		-e DATABASE_PORT=${MONGO_PORT} \
		-e DATABASE_USERNAME=${MONGO_USR} \
		-e DATABASE_PASSWORD=${MONGO_PWD} \
		--network=dev_gorestapi-network \
		--name ${CONTAINER_NAME}\
		${CONTAINER_NAME}:dev

.PHONY: dev-docker-stop
dev-docker-stop:
	${DOCKER} rm -f ${CONTAINER_NAME}

.PHONY: infra-stage
infra-stage:
	#TODO with terraform probably


.PHONY: relocate
relocate:
	@test ${TARGET} || ( echo ">> TARGET is not set. Use: make relocate TARGET=<target>"; exit 1 )
	$(eval ESCAPED_PACKAGENAME := $(shell echo "${PACKAGE_NAME}" | sed -e 's/[\/&]/\\&/g'))
	$(eval ESCAPED_TARGET := $(shell echo "${TARGET}" | sed -e 's/[\/&]/\\&/g'))
	# Renaming package ${PACKAGE_NAME} to ${TARGET}
	@grep -rlI '${PACKAGE_NAME}' * | xargs -I @ ${SED} 's/${ESCAPED_PACKAGENAME}/${ESCAPED_TARGET}/g' @
	# Complete... 
	# NOTE: This does not update the git config nor will it update any imports of the root directory of this project.
