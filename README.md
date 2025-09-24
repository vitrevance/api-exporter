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
