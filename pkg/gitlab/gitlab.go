package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
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
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab: %s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) ListBranches(projectid string) (branches []Branch, err error) {
	u := *c.Base
	u.Path += "/api/v4/projects/" + projectid + "/repository/branches"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab: %s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}
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
