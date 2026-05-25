# api-exporter

_api-exporter_ is a flexible service written in Golang that allows configuring and executing jobs to interact with REST API. It supports multi-step workflows combining scripting, HTTP calls, data transformation, logging, etc, all customizable via YAML configuration.

## Features

- Define multiple named jobs with configurable intervals.
- Execute sequential steps including JavaScript scripts, HTTP requests, field mapping, and printing/logging.
- Use custom transformers to sequence multiple API calls and data processing steps.
- Fetch paginated API data dynamically within JavaScript for complex aggregations.
- Post processed data in a desired JSON or other format to a target API.
- Extensible with various step types for flexible API interactions.

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
- array
- field
- javascript
- parse
- print
- regex
- sequence
- value

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
