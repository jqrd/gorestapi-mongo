WORK IN PROGRESS, DO NOT USE :)

# Base API Example

This API example is a basic framework for a REST API with MongoDB as a database.
It's a fork of [snowzach/gorestapi](https://github.com/snowzach/gorestapi), which supports Postgres as the database.

## Requirements
Protocol Buffers are used to define the db and service schema, and to generate the Go types used for reading/writing data from/to MongoDB. Install `protoc` following the [install docs](https://grpc.io/docs/protoc-installation/).

Docker is required to run MongoDB (and mongo-express for manually inspecting the database).
With Docker installed, set up the required containers with `make infra-dev`.

Other tools are installed automatically when building (or just run `make tools` to install them up front): `protoc-gen-go`, `protoc-gen-gotag`, `mockery`, `swag`.
Ensure that `$HOME/go/bin` is in your path for these external tools to work.

> **Windows support**: sorry, build / infra scripts only works on Linux and MacOS for now.


## Compiling
This requires Go 1.19 or better.

You can clone it anywhere, just run `make` inside the cloned directory to build.

To run the locally built binary, you can use `make run`. That will in turn use `make dev-infra-up` to deploy Mongo and Mongo-Express containers. There's also a launch configuration for VS Code.

To build and run from a container, use `make build-docker`.


## Configuration
The configuration is designed to be specified with environment variables in all caps with underscores instead of periods. 
```
Example:
LOGGER_LEVEL=debug
```

### Options: (TODO reworking configs for MongoDB)
| Setting                         | Description                                                 | Default                 |
| ------------------------------- | ----------------------------------------------------------- | ----------------------- |
| logger.level                    | The default logging level                                   | "info"                  |
| logger.encoding                 | Logging format (console, json or stackdriver)               | "console"               |
| logger.color                    | Enable color in console mode                                | true                    |
| logger.dev_mode                 | Dump additional information as part of log messages         | true                    |
| logger.disable_caller           | Hide the caller source file and line number                 | false                   |
| logger.disable_stacktrace       | Hide a stacktrace on debug logs                             | true                    |
| ---                             | ---                                                         | ---                     |
| metrics.enabled                 | Enable metrics server                                       | true                    |
| metrics.host                    | Host/IP to listen on for metrics server                     | ""                      |
| metrics.port                    | Port to listen on for metrics server                        | 6060                    |
| profiler.enabled                | Enable go profiler on metrics server under /debug/pprof/    | true                    |
| pidfile                         | If set, creates a pidfile at the given path                 | ""                      |
| ---                             | ---                                                         | ---                     |
| server.host                     | The host address to listen on (blank=all addresses)         | ""                      |
| server.port                     | The port number to listen on                                | 8900                    |
| server.tls                      | Enable https/tls                                            | false                   |
| server.devcert                  | Generate a development cert                                 | false                   |
| server.certfile                 | The HTTPS/TLS server certificate                            | "server.crt"            |
| server.keyfile                  | The HTTPS/TLS server key file                               | "server.key"            |
| server.log.enabled              | Log server requests                                         | true                    |
| server.log.level                | Log level for server requests                               | "info                   |
| server.log.request_body         | Log the request body                                        | false                   |
| server.log.response_body        | Log the response body                                       | false                   |
| server.log.ignore_paths         | The endpoint prefixes to not log                            | []string{"/version"}    |
| server.cors.enabled             | Enable CORS middleware                                      | false                   |
| server.cors.allowed_origins     | CORS Allowed origins                                        | []string{"*"}           |
| server.cors.allowed_methods     | CORS Allowed methods                                        | []string{...everything} |
| server.cors.allowed_headers     | CORS Allowed headers                                        | []string{"*"}           |
| server.cors.allowed_credentials | CORS Allowed credentials                                    | false                   |
| server.cors.max_age             | CORS Max Age                                                | 300                     |
| server.metrics.enabled          | Enable metrics on server endpoints                          | true                    |
| server.metrics.ignore_paths     | The endpoint prefixes to not capture metrics on             | []string{"/version"}    |
| ---                             | ---                                                         | ---                     |
| (TODO: Many DB settings are not implemented yet)  |||
| database.username               | The database username                                       | "root"                  |
| database.password               | The database password                                       | "password"              |
| database.host                   | Thos hostname for the database                              | "localhost"             |
| database.port                   | The port for the database                                   | 27017                   |
| database.database               | The database                                                | "gorestapiDB"           |
| database.auto_create            | Automatically create database                               | true                    |
| database.sslmode                | The postgres sslmode to use                                 | "disable"               |
| database.sslcert                | The postgres sslcert file                                   | ""                      |
| database.sslkey                 | The postgres sslkey file                                    | ""                      |
| database.sslrootcert            | The postgres sslrootcert file                               | ""                      |
| database.retries                | How many times to try to reconnect to the database on start | 7                       |
| database.sleep_between_retries  | How long to sleep between retries                           | "7s"                    |
| database.max_connections        | How many pooled connections to have                         | 40                      |
| database.loq_queries            | Log queries (must set logging.level=debug)                  | false                   |
| database.wipe_confirm           | Wipe the database during start                              | false                   |


## Data Storage
Data is stored in a MongoDB database.
In the dev configuration, data is saved in a local `data` directory, which is mounted into the mongo container.


## Query Logic
Find requests like `GET /api/things` and `GET /api/widgets` support a `q` query string JSON argument in the format of a MongoDB filter.
For the documentation on how to use this format see [query documents](https://www.mongodb.com/docs/manual/tutorial/query-documents/) and [query and projection operators](https://www.mongodb.com/docs/manual/reference/operator/query/) in the the MongoDB docs.


## Swagger Documentation
When you run the API it has built in Swagger documentation available at `/api/api-docs/` (trailing slash required)
The documentation is automatically generated.


## TLS/HTTPS
You can enable https by setting the config option server.tls = true and pointing it to your keyfile and certfile.
To create a self-signed cert: `openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 -keyout server.key -out server.crt`
It also has the option to automatically generate a development cert every time it runs using the server.devcert option.


## Relocation
If you want to start with this as boilerplate for your project, you can clone this repo and use the `make relocate` option to rename the package.

```make relocate TARGET=github.com/myname/mycoolproject```
