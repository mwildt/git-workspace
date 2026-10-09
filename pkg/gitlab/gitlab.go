package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	HttpRepoUrl       string `json:"http_url_to_repo"`
	LastActivityAt    string `json:"last_activity_at"`
}

type Branch struct {
	Name   string `json:"name"`
	WebUrl string `json:"web_url"`
	Commit struct {
		Id         string `json:"id"`
		AuthorName string `json:"author_name"`
		CreatedAt  string `json:"created_at"`
	}
}

type Client struct {
	Base  *url.URL
	Token string
	HTTP  *http.Client
}

func (c *Client) ListProjects() (projects []Project, err error) {
	u := *c.Base
	u.Path += "/api/v4/projects"
	q := u.Query()
	q.Set("membership", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		slog.Error("gitlab: failed to build list projects request", "url", u.String(), "error", err)
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.Token)

	slog.Debug("gitlab: listing projects", "url", u.String())
	resp, err := c.HTTP.Do(req)
	if err != nil {
		slog.Error("gitlab: list projects request failed", "url", u.String(), "error", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("gitlab: list projects returned unexpected status", "url", u.String(), "status", resp.Status, "body", string(body))
		return nil, fmt.Errorf("gitlab: %s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		slog.Error("gitlab: failed to decode list projects response", "url", u.String(), "error", err)
		return nil, err
	}
	slog.Debug("gitlab: listed projects", "count", len(projects))
	return projects, nil
}

func (c *Client) ListBranches(projectid string) (branches []Branch, err error) {
	u := *c.Base
	u.Path += "/api/v4/projects/" + projectid + "/repository/branches"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		slog.Error("gitlab: failed to build list branches request", "projectId", projectid, "url", u.String(), "error", err)
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.Token)

	slog.Debug("gitlab: listing branches", "projectId", projectid, "url", u.String())
	resp, err := c.HTTP.Do(req)
	if err != nil {
		slog.Error("gitlab: list branches request failed", "projectId", projectid, "error", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("gitlab: list branches returned unexpected status", "projectId", projectid, "status", resp.Status, "body", string(body))
		return nil, fmt.Errorf("gitlab: %s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		slog.Error("gitlab: failed to decode list branches response", "projectId", projectid, "error", err)
		return nil, err
	}
	slog.Debug("gitlab: listed branches", "projectId", projectid, "count", len(branches))
	return branches, nil
}

func NewClient(token string, baseUrlString string) *Client {
	base, _ := url.Parse(baseUrlString) // oder eigene GitLab-Instanz
	return &Client{
		Base:  base,
		Token: token,
		HTTP:  &http.Client{Timeout: 15 * time.Second},
	}
}
