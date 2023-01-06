# Dockerfile.distroless

ARG EXECUTABLE=gorestapicmd
ARG GIT_VERSION=unknown


FROM golang:1.19.4-bullseye as base

WORKDIR /app/src

COPY ./src/go.mod ./go.mod
COPY ./src/go.sum ./go.sum

RUN go mod download
RUN go mod verify


FROM base as build
ARG EXECUTABLE
ARG GIT_VERSION

WORKDIR /app

COPY ./Makefile ./Makefile
COPY ./src ./src

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build EXECUTABLE=$EXECUTABLE GIT_VERSION=$GIT_VERSION

FROM gcr.io/distroless/static-debian11
ARG EXECUTABLE
ARG GIT_VERSION

ENV EXECUTABLE=${EXECUTABLE}
ENV GIT_VERSION=${GIT_VERSION}

# TODO figure out why all combinations of ENTRYPOINT ..., ENTRYPOINT [], CMD ..., CMD [] lead to /bin/sh still being used and remove this!
COPY --from=busybox:1.35.0-uclibc /bin/sh /bin/sh

WORKDIR /app/
COPY --from=build /app/$EXECUTABLE ./

CMD ./$EXECUTABLE api

EXPOSE 8080