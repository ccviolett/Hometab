package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"hometab/internal/model"
)

func normalizeExternalRequestMethod(method string) (string, error) {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		m = http.MethodGet
	}
	switch m {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return m, nil
	default:
		return "", errors.New("unsupported method")
	}
}

func validateExternalRequestURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url must start with http:// or https://")
	}
	if u.Host == "" || u.Hostname() == "" {
		return errors.New("url host is required")
	}
	if u.User != nil {
		return errors.New("url userinfo is not allowed")
	}
	return nil
}

func validateExternalRequestJSON(raw string, objectOnly bool) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return err
	}
	if objectOnly {
		if _, ok := value.(map[string]interface{}); !ok {
			return errors.New("must be a JSON object")
		}
	}
	return nil
}

func validateParserConfig(parserType, raw string) error {
	switch normalizeDefault(parserType, "status") {
	case "status", "text":
		return validateExternalRequestJSON(raw, false)
	case "json_path":
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil
		}
		var fields []struct {
			Label string `json:"label"`
			Path  string `json:"path"`
		}
		if err := json.Unmarshal([]byte(raw), &fields); err != nil {
			return fmt.Errorf("invalid parser_config_json: %w", err)
		}
		for _, field := range fields {
			if strings.TrimSpace(field.Label) == "" || strings.TrimSpace(field.Path) == "" {
				return errors.New("parser fields require label and path")
			}
			if path := strings.TrimSpace(field.Path); path != "$" && !strings.HasPrefix(path, "$.") {
				return errors.New("parser path must start with $.")
			}
		}
		return nil
	default:
		return errors.New("unsupported parser_type")
	}
}

func normalizeDefault(value, fallback string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return fallback
	}
	return value
}

func normalizeJSONText(raw string) string {
	return strings.TrimSpace(raw)
}

func parseStringMap(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}, nil
	}
	items := make(map[string]string)
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func buildExternalRequestURL(rawURL, queryJSON string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query, err := parseStringMap(queryJSON)
	if err != nil {
		return "", fmt.Errorf("invalid query_json: %w", err)
	}
	values := u.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func buildExternalRequestBody(bodyType, body string) (io.Reader, string) {
	switch normalizeDefault(bodyType, "none") {
	case "json":
		return bytes.NewBufferString(body), "application/json"
	case "form":
		return bytes.NewBufferString(body), "application/x-www-form-urlencoded"
	case "text":
		return bytes.NewBufferString(body), "text/plain; charset=utf-8"
	case "raw":
		return bytes.NewBufferString(body), ""
	default:
		return nil, ""
	}
}

func parseExternalRequestResult(parserType, configJSON string, body []byte, status int, readErr error) []model.ExternalRequestParsedField {
	if readErr != nil {
		return []model.ExternalRequestParsedField{{Label: "读取响应", Error: readErr.Error()}}
	}
	switch normalizeDefault(parserType, "status") {
	case "text":
		value := string(body)
		if len(value) > 500 {
			value = value[:500] + "..."
		}
		return []model.ExternalRequestParsedField{{Label: "响应", Value: value}}
	case "json_path":
		return parseJSONPathFields(configJSON, body)
	default:
		return []model.ExternalRequestParsedField{{Label: "状态", Value: status}}
	}
}

func parseJSONPathFields(configJSON string, body []byte) []model.ExternalRequestParsedField {
	var fields []struct {
		Label string `json:"label"`
		Path  string `json:"path"`
	}
	if strings.TrimSpace(configJSON) == "" {
		fields = []struct {
			Label string `json:"label"`
			Path  string `json:"path"`
		}{{Label: "响应", Path: "$"}}
	} else if err := json.Unmarshal([]byte(configJSON), &fields); err != nil {
		return []model.ExternalRequestParsedField{{Label: "解析配置", Error: err.Error()}}
	}
	var root interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return []model.ExternalRequestParsedField{{Label: "JSON", Error: err.Error()}}
	}
	result := make([]model.ExternalRequestParsedField, 0, len(fields))
	for _, field := range fields {
		value, err := resolveSimpleJSONPath(root, field.Path)
		parsed := model.ExternalRequestParsedField{Label: field.Label, Path: field.Path}
		if err != nil {
			parsed.Error = err.Error()
		} else {
			parsed.Value = value
		}
		result = append(result, parsed)
	}
	return result
}

func resolveSimpleJSONPath(root interface{}, path string) (interface{}, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" {
		return root, nil
	}
	if !strings.HasPrefix(path, "$.") {
		return nil, errors.New("path must start with $.")
	}
	current := root
	for _, token := range strings.Split(strings.TrimPrefix(path, "$."), ".") {
		if token == "" {
			return nil, errors.New("empty path segment")
		}
		name := token
		indexes := make([]int, 0)
		for {
			open := strings.Index(name, "[")
			if open == -1 {
				break
			}
			close := strings.Index(name[open:], "]")
			if close == -1 {
				return nil, errors.New("invalid array index")
			}
			idx, err := strconv.Atoi(name[open+1 : open+close])
			if err != nil {
				return nil, errors.New("invalid array index")
			}
			indexes = append(indexes, idx)
			name = name[:open] + name[open+close+1:]
		}
		if name != "" {
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%s is not an object", name)
			}
			var exists bool
			current, exists = obj[name]
			if !exists {
				return nil, fmt.Errorf("%s not found", name)
			}
		}
		for _, idx := range indexes {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, errors.New("value is not an array")
			}
			if idx < 0 || idx >= len(arr) {
				return nil, errors.New("array index out of range")
			}
			current = arr[idx]
		}
	}
	return current, nil
}
