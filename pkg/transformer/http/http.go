package http

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type HttpTargetConfig struct {
	// Required request URL (no default)
	URL string `yaml:"url"`

	// HTTP method (default GET)
	Method string `yaml:"method"`

	// Request headers
	Headers map[string]string `yaml:"headers"`

	// URL query parameters
	QueryParams map[string]string `yaml:"query_params"`

	// Request body as bytes
	Body string `yaml:"body"`

	// Basic auth credentials
	BasicAuthUsername string `yaml:"basic_auth_username"`
	BasicAuthPassword string `yaml:"basic_auth_password"`

	// Client-level options (with ms timeouts)
	TimeoutMillis         int    `yaml:"timeout_ms"`
	IdleConnTimeoutMillis int    `yaml:"idle_conn_timeout_ms"`
	MaxIdleConns          int    `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int    `yaml:"max_idle_conns_per_host"`
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify"`
	ProxyURL              string `yaml:"proxy_url"`

	// Follow redirects (default true)
	FollowRedirects *bool `yaml:"follow_redirects"`
}

func (c *HttpTargetConfig) MergeMap(cfg map[string]any) error {
	for key, value := range cfg {
		switch key {
		case "url":
			if v, ok := value.(string); ok {
				c.URL = v
			} else {
				return fmt.Errorf("invalid type for url, expected string")
			}
		case "method":
			if v, ok := value.(string); ok {
				c.Method = v
			} else {
				return fmt.Errorf("invalid type for method, expected string")
			}
		case "headers":
			if m, ok := value.(map[string]any); ok {
				if c.Headers == nil {
					c.Headers = make(map[string]string)
				}
				for hk, hv := range m {
					if hs, ok := hv.(string); ok {
						c.Headers[hk] = hs
					} else {
						return fmt.Errorf("invalid header value type for %s, expected string", hk)
					}
				}
			} else {
				return fmt.Errorf("invalid type for headers, expected map[string]any")
			}
		case "query_params":
			if m, ok := value.(map[string]any); ok {
				if c.QueryParams == nil {
					c.QueryParams = make(map[string]string)
				}
				for qk, qv := range m {
					if qs, ok := qv.(string); ok {
						c.QueryParams[qk] = qs
					} else {
						return fmt.Errorf("invalid query param value type for %s, expected string", qk)
					}
				}
			} else {
				return fmt.Errorf("invalid type for query_params, expected map[string]any")
			}
		case "body":
			// Accept string or []byte for simplicity
			switch v := value.(type) {
			case string:
				c.Body = v
			default:
				return fmt.Errorf("invalid type for body, expected string or []byte")
			}
		case "basic_auth_username":
			if v, ok := value.(string); ok {
				c.BasicAuthUsername = v
			} else {
				return fmt.Errorf("invalid type for basic_auth_username, expected string")
			}
		case "basic_auth_password":
			if v, ok := value.(string); ok {
				c.BasicAuthPassword = v
			} else {
				return fmt.Errorf("invalid type for basic_auth_password, expected string")
			}
		case "timeout_ms":
			if iv, err := toInt(value); err == nil {
				c.TimeoutMillis = iv
			} else {
				return fmt.Errorf("invalid type for timeout_ms: %v", err)
			}
		case "idle_conn_timeout_ms":
			if iv, err := toInt(value); err == nil {
				c.IdleConnTimeoutMillis = iv
			} else {
				return fmt.Errorf("invalid type for idle_conn_timeout_ms: %v", err)
			}
		case "max_idle_conns":
			if iv, err := toInt(value); err == nil {
				c.MaxIdleConns = iv
			} else {
				return fmt.Errorf("invalid type for max_idle_conns: %v", err)
			}
		case "max_idle_conns_per_host":
			if iv, err := toInt(value); err == nil {
				c.MaxIdleConnsPerHost = iv
			} else {
				return fmt.Errorf("invalid type for max_idle_conns_per_host: %v", err)
			}
		case "tls_insecure_skip_verify":
			if bv, ok := value.(bool); ok {
				c.TLSInsecureSkipVerify = bv
			} else {
				return fmt.Errorf("invalid type for tls_insecure_skip_verify, expected bool")
			}
		case "proxy_url":
			if v, ok := value.(string); ok {
				c.ProxyURL = v
			} else {
				return fmt.Errorf("invalid type for proxy_url, expected string")
			}
		case "follow_redirects":
			if bv, ok := value.(bool); ok {
				c.FollowRedirects = &bv
			} else {
				return fmt.Errorf("invalid type for follow_redirects, expected bool")
			}
		}
	}
	return nil
}

// toInt attempts to convert value (int, float64, string) to int
func toInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, errors.New("cannot convert to int: " + reflect.TypeOf(value).String())
	}
}

// NewHttpTargetConfig returns a config with reasonable defaults except URL
func NewHttpTargetConfig() *HttpTargetConfig {
	trueVal := true
	return &HttpTargetConfig{
		Method:                http.MethodGet,
		Headers:               make(map[string]string),
		QueryParams:           make(map[string]string),
		TimeoutMillis:         30000, // 30 seconds
		IdleConnTimeoutMillis: 90000, // 90 seconds
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		TLSInsecureSkipVerify: false,
		FollowRedirects:       &trueVal, // default to follow redirects
	}
}

// CreateHttpClient builds *http.Client from config
func (c *HttpTargetConfig) CreateHttpClient() (*http.Client, error) {
	timeout := time.Duration(c.TimeoutMillis) * time.Millisecond
	tlsSkip := c.TLSInsecureSkipVerify

	transport := &http.Transport{
		IdleConnTimeout:     time.Duration(c.IdleConnTimeoutMillis) * time.Millisecond,
		MaxIdleConns:        c.MaxIdleConns,
		MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsSkip,
		},
	}
	if c.ProxyURL != "" {
		proxyURL, err := url.Parse(c.ProxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	if c.FollowRedirects != nil && !*c.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		client.CheckRedirect = nil
	}

	return client, nil
}

// CreateHttpRequest builds *http.Request with configured properties
func (c *HttpTargetConfig) CreateHttpRequest() (*http.Request, error) {
	if c.URL == "" {
		return nil, errors.New("url must be specified")
	}
	method := c.Method
	if method == "" {
		method = http.MethodGet
	}

	reqURL, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}
	q := reqURL.Query()
	for key, val := range c.QueryParams {
		q.Set(key, val)
	}
	reqURL.RawQuery = q.Encode()

	var bodyReader io.Reader
	if len(c.Body) > 0 {
		bodyReader = bytes.NewReader([]byte(c.Body))
	}

	req, err := http.NewRequest(method, reqURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	for key, val := range c.Headers {
		req.Header.Set(key, val)
	}

	if c.BasicAuthUsername != "" {
		req.SetBasicAuth(c.BasicAuthUsername, c.BasicAuthPassword)
	}

	return req, nil
}

func init() {
	transformer.RegisterTransformerFactory("http", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := NewHttpTargetConfig()
		err := value.Decode(t)
		return &httpTransformer{
			Config: t,
		}, err
	}))
}

type httpTransformer struct {
	Config *HttpTargetConfig
}

func (this *httpTransformer) Transform(ctx *transformer.TransformationContext) error {
	mp, ok := ctx.Object.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid source")
	}

	this.Config.MergeMap(mp)

	client, err := this.Config.CreateHttpClient()
	if err != nil {
		return err
	}
	req, err := this.Config.CreateHttpRequest()
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	obj, err := responseToMap(resp)
	if err != nil {
		return err
	}
	ctx.Result = obj

	return nil
}

func responseToMap(resp *http.Response) (map[string]any, error) {
	bodyBytes := []byte{}
	if resp.Body != nil {

		defer resp.Body.Close()

		// Read full response body
		var err error
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	// Convert headers to map[string]string (joining multiple values by comma)
	headers := make(map[string]string)
	for k, vals := range resp.Header {
		// Join multiple values with comma as per RFC 7230 section 3.2.2
		joined := ""
		for i, v := range vals {
			if i > 0 {
				joined += ", "
			}
			joined += v
		}
		headers[k] = joined
	}

	result := map[string]any{
		"status":      resp.Status,
		"status_code": resp.StatusCode,
		"headers":     headers,
		"body":        bodyBytes,
	}

	return result, nil
}
