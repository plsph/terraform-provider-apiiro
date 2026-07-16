package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type apiPagedResponse[T any] struct {
	Items  []T `json:"items"`
	Paging struct {
		PageSize       int `json:"pageSize"`
		Skip           int `json:"skip"`
		TotalItemCount int `json:"totalItemCount"`
	} `json:"paging"`
}

type scmRepository struct {
	ApproximateFileCount          *int     `json:"approximateFileCount"`
	ApproximateSizeMB             *int     `json:"approximateSizeMb"`
	AssociatedProjectIDs          []string `json:"associatedProjectIds"`
	CreatedAt                     *string  `json:"createdAt"`
	DefaultBranch                 *string  `json:"defaultBranch"`
	HTTPCloneURL                  *string  `json:"httpCloneUrl"`
	ID                            *string  `json:"id"`
	IgnoredBy                     *string  `json:"ignoredBy"`
	IgnoreReason                  *string  `json:"ignoreReason"`
	IsArchived                    bool     `json:"isArchived"`
	IsIgnored                     bool     `json:"isIgnored"`
	IsPublic                      bool     `json:"isPublic"`
	Key                           *string  `json:"key"`
	Languages                     []string `json:"languages"`
	LastMonitoringChangeTimestamp *string  `json:"lastMonitoringChangeTimestamp"`
	MonitoredBranches             []string `json:"monitoredBranches"`
	MonitorStatus                 *string  `json:"monitorStatus"`
	Name                          *string  `json:"name"`
	ProjectExternalID             *string  `json:"projectExternalId"`
	ProjectID                     *string  `json:"projectId"`
	ProjectURL                    *string  `json:"projectUrl"`
	Provider                      *string  `json:"provider"`
	RepositoryExternalID          *string  `json:"repositoryExternalId"`
	ServerURL                     *string  `json:"serverUrl"`
	SSHURL                        *string  `json:"sshUrl"`
	URL                           *string  `json:"url"`
}

type tagBody struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type repositoryTagResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func applyAPIFilters(query url.Values, filters map[string][]string) {
	for key, values := range filters {
		name := strings.TrimSpace(key)
		if name == "" {
			continue
		}
		for _, value := range values {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			query.Add(fmt.Sprintf("filters[%s]", name), trimmed)
		}
	}
}

func (c *Client) listRoleGroups(ctx context.Context) ([]roleGroupReference, error) {
	const pageSize = 100
	all := make([]roleGroupReference, 0)
	skip := 0

	for {
		endpoint := fmt.Sprintf("/rest-api/v1/roleGroups?skip=%d&pageSize=%d", skip, pageSize)
		var out apiPagedResponse[roleGroupReference]
		if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Items...)

		skip += len(out.Items)
		if len(out.Items) == 0 || skip >= out.Paging.TotalItemCount {
			break
		}
	}
	return all, nil
}

func (c *Client) getRoleGroup(ctx context.Context, key string) (*roleGroupBody, error) {
	var out roleGroupBody
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/rest-api/v1/roleGroups/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) createRoleGroup(ctx context.Context, body roleGroupBody) (string, error) {
	var out string
	if err := c.doJSON(ctx, http.MethodPost, "/rest-api/v1/roleGroups", body, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) updateRoleGroup(ctx context.Context, key string, body roleGroupBody) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/rest-api/v1/roleGroups/%s", pathEscape(key)), body, nil)
}

func (c *Client) deleteRoleGroup(ctx context.Context, key string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/rest-api/v1/roleGroups/%s", pathEscape(key)), nil, nil)
}

type monitorBranchBody struct {
	BranchName string `json:"branchName"`
}

type problemDetails struct {
	Message string `json:"message"`
}

func NewClient(baseURL, token string) (*Client, error) {
	baseURL = strings.TrimSpace(strings.TrimSuffix(baseURL, "/"))
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	return &Client{
		baseURL: baseURL,
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (c *Client) listScmRepositories(ctx context.Context, filters map[string][]string) ([]scmRepository, error) {
	const pageSize = 100
	all := make([]scmRepository, 0)
	skip := 0

	for {
		query := url.Values{}
		query.Set("skip", fmt.Sprintf("%d", skip))
		query.Set("pageSize", fmt.Sprintf("%d", pageSize))
		applyAPIFilters(query, filters)
		endpoint := "/rest-api/v1/ScmRepositories?" + query.Encode()
		var out apiPagedResponse[scmRepository]
		if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Items...)

		skip += len(out.Items)
		if len(out.Items) == 0 || skip >= out.Paging.TotalItemCount {
			break
		}
	}
	return all, nil
}

func (c *Client) listRepositoriesV2(ctx context.Context, filters map[string][]string) ([]repositoryBodyV2, error) {
	const pageSize = 1000
	all := make([]repositoryBodyV2, 0)
	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	applyAPIFilters(query, filters)

	for {
		endpoint := "/rest-api/v2/repositories"
		if encoded := query.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}

		var out struct {
			Next  json.RawMessage    `json:"next"`
			Items []repositoryBodyV2 `json:"items"`
		}
		if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Items...)

		rawNext := bytes.TrimSpace(out.Next)
		if len(rawNext) == 0 || string(rawNext) == "null" {
			break
		}

		query = url.Values{}
		query.Set("pageSize", fmt.Sprintf("%d", pageSize))
		applyAPIFilters(query, filters)

		var nextMap map[string]string
		if err := json.Unmarshal(rawNext, &nextMap); err == nil {
			hasNext := false
			for key, value := range nextMap {
				if strings.TrimSpace(value) == "" {
					continue
				}
				query.Set(key, value)
				hasNext = true
			}
			if !hasNext {
				break
			}
			continue
		}

		var nextToken string
		if err := json.Unmarshal(rawNext, &nextToken); err == nil {
			nextToken = strings.TrimSpace(nextToken)
			if nextToken == "" {
				break
			}
			query.Set("next", nextToken)
			continue
		}

		return nil, fmt.Errorf("unexpected repositories pagination token format: %s", string(rawNext))
	}

	return all, nil
}

func (c *Client) getScmRepositoryByKey(ctx context.Context, repositoryKey string) (*scmRepository, error) {
	trimmedKey := strings.TrimSpace(repositoryKey)
	if trimmedKey == "" {
		return nil, nil
	}

	for _, nameHint := range scmRepositoryNameCandidates(trimmedKey) {
		filtered, err := c.listScmRepositories(ctx, map[string][]string{"RepositoryName": {nameHint}})
		if err != nil {
			return nil, err
		}
		if repo := findScmRepositoryByKey(filtered, trimmedKey); repo != nil {
			return repo, nil
		}
	}

	repositories, err := c.listScmRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	if repo := findScmRepositoryByKey(repositories, trimmedKey); repo != nil {
		return repo, nil
	}
	return nil, nil
}

func findScmRepositoryByKey(repositories []scmRepository, repositoryKey string) *scmRepository {
	for i := range repositories {
		if repositories[i].Key != nil && *repositories[i].Key == repositoryKey {
			return &repositories[i]
		}
	}
	return nil
}

func scmRepositoryNameCandidates(repositoryKey string) []string {
	candidates := make([]string, 0, 3)
	seen := map[string]struct{}{}

	add := func(raw string) {
		value := strings.TrimSpace(raw)
		if value == "" {
			return
		}
		value = strings.TrimSuffix(value, ".git")
		if value == "" {
			return
		}
		if _, exists := seen[value]; exists {
			return
		}
		seen[value] = struct{}{}
		candidates = append(candidates, value)
	}

	if parsed, err := url.Parse(repositoryKey); err == nil {
		if segment := path.Base(strings.TrimSpace(parsed.Path)); segment != "." && segment != "/" {
			add(segment)
		}
	}

	if idx := strings.LastIndex(repositoryKey, "/"); idx >= 0 && idx+1 < len(repositoryKey) {
		add(repositoryKey[idx+1:])
	}

	parts := strings.Split(repositoryKey, "_")
	if len(parts) > 0 {
		add(parts[len(parts)-1])
	}

	return candidates
}

func (c *Client) monitorRepository(ctx context.Context, repositoryKey string) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/monitor", pathEscape(repositoryKey)), nil, nil)
}

func (c *Client) unmonitorRepository(ctx context.Context, repositoryKey string) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/unmonitor", pathEscape(repositoryKey)), nil, nil)
}

func (c *Client) monitorBranch(ctx context.Context, repositoryKey, branch string) error {
	payload := monitorBranchBody{BranchName: branch}
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/monitorBranch", pathEscape(repositoryKey)), payload, nil)
}

func (c *Client) unmonitorBranch(ctx context.Context, repositoryKey, branch string) error {
	payload := monitorBranchBody{BranchName: branch}
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/unmonitorBranch", pathEscape(repositoryKey)), payload, nil)
}

func (c *Client) listRepositoryTags(ctx context.Context, repositoryKey string) ([]repositoryTagResponse, error) {
	var tags []repositoryTagResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/tags", pathEscape(repositoryKey)), nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *Client) upsertRepositoryTag(ctx context.Context, repositoryKey, name, value string) error {
	payload := tagBody{Name: name, Value: value}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/tags", pathEscape(repositoryKey)), payload, nil)
}

func (c *Client) deleteRepositoryTag(ctx context.Context, repositoryKey, tagName string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/tags/%s", pathEscape(repositoryKey), pathEscape(tagName)), nil, nil)
}

func (c *Client) deleteEngagement(ctx context.Context, key string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/rest-api/v1/engagements/%s", pathEscape(key)), nil, nil)
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, body any, out any) error {
	fullURL := strings.TrimSuffix(c.baseURL, "/") + endpoint

	var bodyReader io.Reader
	var reqBody []byte
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = raw
		bodyReader = bytes.NewBuffer(raw)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	tflog.Debug(ctx, "api request", map[string]any{
		"method":  method,
		"url":     fullURL,
		"headers": sanitizeHeaders(req.Header),
		"body":    string(reqBody),
	})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respData, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	tflog.Debug(ctx, "api response", map[string]any{
		"method":  method,
		"url":     fullURL,
		"status":  resp.StatusCode,
		"headers": sanitizeHeaders(resp.Header),
		"body":    string(respData),
	})

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		var details problemDetails
		_ = json.Unmarshal(respData, &details)
		if details.Message != "" {
			return fmt.Errorf("api error: status %d: %s", resp.StatusCode, details.Message)
		}
		if len(respData) > 0 {
			return fmt.Errorf("api error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respData)))
		}
		return fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(respData, out); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func sanitizeHeaders(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for key, values := range headers {
		joined := strings.Join(values, ",")
		if strings.EqualFold(key, "Authorization") {
			out[key] = "<redacted>"
			continue
		}
		out[key] = joined
	}
	return out
}

func pathEscape(v string) string {
	return url.PathEscape(v)
}
