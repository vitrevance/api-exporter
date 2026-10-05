# api-exporter

_api-exporter_ is a flexible service written in Golang that allows configuring and executing jobs to interact with REST API. It supports multi-step workflows combining scripting, HTTP calls, data transformation, logging, etc, all customizable via YAML configuration.

## Features

- Define multiple named jobs with configurable intervals.
- Execute sequential steps including JavaScript scripts, HTTP requests, field mapping, and printing/logging.
- Use custom transformers to sequence multiple API calls and data processing steps.
- Fetch paginated API data dynamically within JavaScript for complex aggregations.
- Post processed data in a desired JSON or other format to a target API.
- Extensible with various step types for flexible API interactions.

## Command-line flags

```text
-config string
      Path or HTTP(S) URL of the YAML configuration file (default "config.yaml")
-reloadInterval duration
      How often to check and reload the configuration; 0s loads it once (default "0s")
-addr string
      HTTP server listen address, such as :8080; when omitted, only jobs are run
-docker
      Enable Docker access for JavaScript transformers using Docker client features
```

For example, start the HTTP server on port 8080 and check for configuration
updates every 30 seconds:

```bash
api-exporter -config ./config.yaml -addr :8080 -reloadInterval 30s
```

## Example

Fetch data from `https://api.example.com/data` and push it as-is to `https://api.target.com/submit`

```yaml
jobs:
  - job_name: example-job
    interval: 30s
    steps:
      - type: http
        url: https://api.example.com/data
        method: GET
      - type: field
        source: body
        target: body
        map:
          type: parse
          format: from_bytes
      - type: print
        format: 'Received data: %v'
        log: true
      - type: http
        url: https://api.target.com/submit
        method: POST
        headers:
          Content-Type: application/json
      - type: print
        format: 'Submission status: %v'
        log: true
```

## Transformation types

- http
- file
- array
- field
- javascript
- parse
- print
- regex
- sequence
- value

### File transformer

Use `operation: read` to load a file. Like the HTTP transformer, the result is
a map whose `body` field contains binary data (`[]byte`):

```yaml
- type: file
  operation: read
  path: ./input.json
```

Use `operation: write` to write the current binary or string value. If the
current value is a map (for example, an HTTP response), its `body` field is
written:

```yaml
- type: file
  operation: write
  path: ./output.bin
```

Both `operation` and `path` can also be supplied by fields on the current
transformation object; those values override the YAML configuration for that
invocation.

## Transformers as HTTP Handlers

Transformers can also be used as HTTP handlers. When a request is made to the server with a URL path that matches a transformer name, the transformer will be executed with the request data as input.

The request data includes:
- `method`: The HTTP method of the request
- `url`: The full URL of the request
- `headers`: The headers of the request
- `body`: The body of the request (if present)
- `query`: The query parameters of the request

The transformer can access this data through the `source` object in its script.

Example configuration:

```yaml
transformers:
  ok:
    type: javascript
    script: |
      return {body: "hello world!!! " + source.url, status_code: 200};
```

In this example, a request to `/ok` will execute the `ok` transformer, which returns a response with the body "hello world!!! " concatenated with the request URL and a status code of 200.

# Docker Client Setup

The exporter can interact with a Docker daemon, for example to inspect containers or manage images within job steps. To enable this functionality, run the exporter with the `-docker` flag.

When enabled, the Docker client is initialized using `client.FromEnv`, which automatically reads the standard Docker environment variables. This allows you to connect to remote Docker hosts or configure TLS securely without code changes.

Set the following environment variables as needed:

- `DOCKER_HOST`: The URL of the Docker daemon (e.g., `tcp://my-docker-host:2375` or `unix:///var/run/docker.sock`).
- `DOCKER_TLS_VERIFY`: Set to `1` to enable TLS verification for secure communication with the daemon.
- `DOCKER_CERT_PATH`: Path to the directory containing TLS certificates (`ca.pem`, `cert.pem`, `key.pem`) used for authentication.

## Example Usage

```bash
DOCKER_HOST=unix:///var/run/docker.sock ./api-exporter -config=config.yaml -docker
```

### Docker functions in JavaScript

When Docker support is enabled, JavaScript transformers can list, inspect, and
stop containers with these functions:

- `docker_container_list(filters, all, limit)` returns the matching containers.
  `filters` maps Docker filter names to arrays of accepted values, `all` includes
  stopped containers, and `limit` sets the maximum number of results. Use `0`
  for no limit.
- `docker_container_inspect(id)` returns detailed Docker API information for the
  container identified by its ID or name.
- `docker_container_kill(id, signal)` sends a signal such as `SIGTERM` or
  `SIGKILL` to the container identified by its ID or name. It returns `null` on
  success.

All three functions return an object with an `error` field if Docker support is
not enabled or the Docker API request fails.

For example, this transformer finds running containers with the Compose service
label `worker` and sends `SIGTERM` to each one:

```yaml
transformers:
  stop-workers:
    type: javascript
    script: |
      var containers = docker_container_list(
        {"label": ["com.docker.compose.service=worker"]},
        false,
        0
      );
      if (containers.error) {
        return containers;
      }

      for (var i = 0; i < containers.length; i++) {
        var result = docker_container_kill(containers[i].Id, "SIGTERM");
        if (result && result.error) {
          return result;
        }
      }

      return {stopped: containers.length};
```
