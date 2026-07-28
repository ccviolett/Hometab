package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
)

const externalRequestMaxBodyBytes int64 = 1024 * 1024

type ExternalRequestSvc struct {
	repo             *repository.ExternalRequestRepo
	http             *externalRequestHTTP
	executionAllowed bool
}

func NewExternalRequestSvc(repo *repository.ExternalRequestRepo, executionAllowed ...bool) *ExternalRequestSvc {
	allowed := true
	if len(executionAllowed) > 0 {
		allowed = executionAllowed[0]
	}
	return &ExternalRequestSvc{repo: repo, http: newExternalRequestHTTP(), executionAllowed: allowed}
}

func (s *ExternalRequestSvc) FindAll() ([]model.ExternalRequest, error) {
	return s.repo.FindAll()
}

func (s *ExternalRequestSvc) FindByID(id uuid.UUID) (*model.ExternalRequest, error) {
	return s.repo.FindByID(id)
}

func (s *ExternalRequestSvc) Create(req model.ExternalRequestCreate) (*model.ExternalRequest, error) {
	method, err := normalizeExternalRequestMethod(req.Method)
	if err != nil {
		return nil, err
	}
	if err := validateExternalRequestURL(req.URL); err != nil {
		return nil, err
	}
	bodyType := normalizeDefault(req.BodyType, "none")
	if bodyType == "json" && strings.TrimSpace(req.Body) != "" {
		if err := validateExternalRequestJSON(req.Body, false); err != nil {
			return nil, fmt.Errorf("invalid json body: %w", err)
		}
	}
	parserType := normalizeDefault(req.ParserType, "status")
	if err := validateExternalRequestJSON(req.HeadersJSON, true); err != nil {
		return nil, fmt.Errorf("invalid headers_json: %w", err)
	}
	if err := validateExternalRequestJSON(req.QueryJSON, true); err != nil {
		return nil, fmt.Errorf("invalid query_json: %w", err)
	}
	if err := validateParserConfig(parserType, req.ParserConfigJSON); err != nil {
		return nil, err
	}
	confirm := method != http.MethodGet
	if req.ConfirmBeforeRun != nil {
		confirm = *req.ConfirmBeforeRun
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	item := model.ExternalRequest{
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		Method:           method,
		URL:              strings.TrimSpace(req.URL),
		HeadersJSON:      normalizeJSONText(req.HeadersJSON),
		QueryJSON:        normalizeJSONText(req.QueryJSON),
		BodyType:         bodyType,
		Body:             req.Body,
		ParserType:       parserType,
		ParserConfigJSON: normalizeJSONText(req.ParserConfigJSON),
		ConfirmBeforeRun: confirm,
		Enabled:          enabled,
		OrderIndex:       req.OrderIndex,
	}
	if item.Name == "" || item.URL == "" {
		return nil, errors.New("name and url are required")
	}
	if err := s.repo.Create(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ExternalRequestSvc) Update(id uuid.UUID, req model.ExternalRequestUpdate) (*model.ExternalRequest, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		item.Name = strings.TrimSpace(*req.Name)
		if item.Name == "" {
			return nil, errors.New("name is required")
		}
	}
	if req.Description != nil {
		item.Description = strings.TrimSpace(*req.Description)
	}
	if req.Method != nil {
		method, err := normalizeExternalRequestMethod(*req.Method)
		if err != nil {
			return nil, err
		}
		item.Method = method
	}
	if req.URL != nil {
		if err := validateExternalRequestURL(*req.URL); err != nil {
			return nil, err
		}
		item.URL = strings.TrimSpace(*req.URL)
	}
	if req.HeadersJSON != nil {
		if err := validateExternalRequestJSON(*req.HeadersJSON, true); err != nil {
			return nil, fmt.Errorf("invalid headers_json: %w", err)
		}
		item.HeadersJSON = normalizeJSONText(*req.HeadersJSON)
	}
	if req.QueryJSON != nil {
		if err := validateExternalRequestJSON(*req.QueryJSON, true); err != nil {
			return nil, fmt.Errorf("invalid query_json: %w", err)
		}
		item.QueryJSON = normalizeJSONText(*req.QueryJSON)
	}
	if req.BodyType != nil {
		item.BodyType = normalizeDefault(*req.BodyType, "none")
	}
	if req.Body != nil {
		item.Body = *req.Body
	}
	if item.BodyType == "json" && strings.TrimSpace(item.Body) != "" {
		if err := validateExternalRequestJSON(item.Body, false); err != nil {
			return nil, fmt.Errorf("invalid json body: %w", err)
		}
	}
	if req.ParserType != nil {
		if err := validateParserConfig(*req.ParserType, item.ParserConfigJSON); err != nil {
			return nil, err
		}
		item.ParserType = normalizeDefault(*req.ParserType, "status")
	}
	if req.ParserConfigJSON != nil {
		if err := validateParserConfig(item.ParserType, *req.ParserConfigJSON); err != nil {
			return nil, err
		}
		item.ParserConfigJSON = normalizeJSONText(*req.ParserConfigJSON)
	}
	if req.ConfirmBeforeRun != nil {
		item.ConfirmBeforeRun = *req.ConfirmBeforeRun
	}
	if req.Enabled != nil {
		item.Enabled = *req.Enabled
	}
	if req.OrderIndex != nil {
		item.OrderIndex = *req.OrderIndex
	}
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ExternalRequestSvc) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *ExternalRequestSvc) Execute(id uuid.UUID) (*model.ExternalRequestExecuteResult, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !item.Enabled {
		return nil, errors.New("request is disabled")
	}
	if !s.executionAllowed {
		return nil, errors.New("execution_disabled")
	}

	requestURL, err := buildExternalRequestURL(item.URL, item.QueryJSON)
	if err != nil {
		return nil, err
	}
	body, contentType := buildExternalRequestBody(item.BodyType, item.Body)
	req, err := http.NewRequest(item.Method, requestURL, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	headers, err := parseStringMap(item.HeadersJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid headers_json: %w", err)
	}
	for key, value := range headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := s.http.do(req)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		code := externalRequestError(err)
		return &model.ExternalRequestExecuteResult{
			DurationMS: duration,
			Parsed:     []model.ExternalRequestParsedField{{Label: "错误", Error: code}},
			Error:      code,
		}, nil
	}
	defer resp.Body.Close()

	data, readErr := io.ReadAll(io.LimitReader(resp.Body, externalRequestMaxBodyBytes+1))
	tooLarge := int64(len(data)) > externalRequestMaxBodyBytes
	if tooLarge {
		data = data[:externalRequestMaxBodyBytes]
	}
	bodyPreview := string(data)
	if len(bodyPreview) > 4096 {
		bodyPreview = bodyPreview[:4096]
	}
	if tooLarge {
		bodyPreview += "\n...[response limit exceeded]"
	}
	result := &model.ExternalRequestExecuteResult{
		Status:      resp.StatusCode,
		StatusText:  resp.Status,
		DurationMS:  duration,
		Headers:     resp.Header,
		BodyPreview: bodyPreview,
		Parsed:      parseExternalRequestResult(item.ParserType, item.ParserConfigJSON, data, resp.StatusCode, readErr),
	}
	if readErr != nil {
		result.Error = "upstream_error"
	}
	if tooLarge {
		result.Error = "response_too_large"
		result.Parsed = []model.ExternalRequestParsedField{{Label: "错误", Error: result.Error}}
	}
	return result, nil
}
