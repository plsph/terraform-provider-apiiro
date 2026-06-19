package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func (c *Client) listScmRepositories() ([]scmRepository, error) {
	const pageSize = 100
	all := make([]scmRepository, 0)
	skip := 0

	for {
		endpoint := fmt.Sprintf("/rest-api/v1/ScmRepositories?skip=%d&pageSize=%d", skip, pageSize)
		var out apiPagedResponse[scmRepository]
		if err := c.doJSON(http.MethodGet, endpoint, nil, &out); err != nil {
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

func (c *Client) getScmRepositoryByKey(repositoryKey string) (*scmRepository, error) {
	repositories, err := c.listScmRepositories()
	if err != nil {
		return nil, err
	}
	for i := range repositories {
		if repositories[i].Key != nil && *repositories[i].Key == repositoryKey {
			return &repositories[i], nil
		}
	}
	return nil, nil
}

func (c *Client) monitorRepository(repositoryKey string) error {
	return c.doJSON(http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/monitor", pathEscape(repositoryKey)), nil, nil)
}

func (c *Client) unmonitorRepository(repositoryKey string) error {
	return c.doJSON(http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/unmonitor", pathEscape(repositoryKey)), nil, nil)
}

func (c *Client) monitorBranch(repositoryKey, branch string) error {
	payload := monitorBranchBody{BranchName: branch}
	return c.doJSON(http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/monitorBranch", pathEscape(repositoryKey)), payload, nil)
}

func (c *Client) unmonitorBranch(repositoryKey, branch string) error {
	payload := monitorBranchBody{BranchName: branch}
	return c.doJSON(http.MethodPut, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/unmonitorBranch", pathEscape(repositoryKey)), payload, nil)
}

func (c *Client) listRepositoryTags(repositoryKey string) ([]repositoryTagResponse, error) {
	var tags []repositoryTagResponse
	err := c.doJSON(http.MethodGet, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/tags", pathEscape(repositoryKey)), nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *Client) upsertRepositoryTag(repositoryKey, name, value string) error {
	payload := tagBody{Name: name, Value: value}
	return c.doJSON(http.MethodPost, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/tags", pathEscape(repositoryKey)), payload, nil)
}

func (c *Client) deleteRepositoryTag(repositoryKey, tagName string) error {
	return c.doJSON(http.MethodDelete, fmt.Sprintf("/rest-api/v1/ScmRepositories/%s/tags/%s", pathEscape(repositoryKey), pathEscape(tagName)), nil, nil)
}

func (c *Client) doJSON(method, endpoint string, body any, out any) error {
	fullURL := strings.TrimSuffix(c.baseURL, "/") + endpoint

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		var details problemDetails
		_ = json.Unmarshal(data, &details)
		if details.Message != "" {
			return fmt.Errorf("api error: status %d: %s", resp.StatusCode, details.Message)
		}
		if len(data) > 0 {
			return fmt.Errorf("api error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
		}
		return fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func pathEscape(v string) string {
	return url.PathEscape(v)
}
