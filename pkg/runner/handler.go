package runner

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/vitrevance/api-exporter/pkg/transformer"
)

func (this *Config) Handle(w http.ResponseWriter, r *http.Request) {
	// Get the first segment of the path
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathSegments) > 0 {
		firstSegment := pathSegments[0]

		// Check if there's a transformer with this name
		if tr, exists := this.Transformers[firstSegment]; exists {
			// Prepare transformation context with request data
			requestData := map[string]any{
				"method":  r.Method,
				"url":     r.URL.String(),
				"headers": r.Header,
			}

			// Read request body if present
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					requestData["body"] = string(bodyBytes)
				}
			}

			// Add query parameters
			queryParams := make(map[string][]string)
			for key, values := range r.URL.Query() {
				if len(values) > 0 {
					queryParams[key] = values
				}
			}
			requestData["query"] = queryParams

			// Create transformation context
			tctx := &transformer.TransformationContext{
				Object:       requestData,
				Result:       make(map[string]any),
				Transformers: this.Transformers,
			}

			// Run transformation
			err := tr.Transform(tctx)
			if err != nil {
				http.Error(w, "Transformation error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Extract HTTP response properties from Result if it's a map

			if resultMap, ok := tctx.Result.(map[string]any); ok {
				statusCode := http.StatusOK
				headers := make(map[string]string)
				// Extract status code if present
				if statusCodeVal, exists := resultMap["status_code"]; exists {
					switch v := statusCodeVal.(type) {
					case int:
						statusCode = v
					case float64:
						statusCode = int(v)
					case float32:
						statusCode = int(v)
					case int64:
						statusCode = int(v)
					case int32:
						statusCode = int(v)
					case int16:
						statusCode = int(v)
					case int8:
						statusCode = int(v)
					}
				}

				// Extract headers if present
				if headersMap, exists := resultMap["headers"]; exists {
					if headerMap, ok := headersMap.(map[string]any); ok {
						for key, value := range headerMap {
							if stringValue, ok := value.(string); ok {
								headers[key] = stringValue
							}
						}
					}
				}

				// Use body if present, otherwise use the whole result
				if bodyVal, exists := resultMap["body"]; exists {
					// Set headers
					for key, value := range headers {
						w.Header().Set(key, value)
					}

					w.WriteHeader(statusCode)

					if bodyString, ok := bodyVal.(string); ok {
						if _, err := w.Write([]byte(bodyString)); err != nil {
							http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
						}
						return
					} else {
						if err := json.NewEncoder(w).Encode(bodyVal); err != nil {
							http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
						}
						return
					}
				}
			}
			http.Error(w, "Failed to process response", http.StatusInternalServerError)
			return
		}
	}
	http.Error(w, "Not found", http.StatusNotFound)
}
