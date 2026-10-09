package gitlab

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
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

func NewTLSConfig(rootCaFile string, rootCaDir string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	if rootCaFile != "" {
		pem, err := os.ReadFile(rootCaFile)
		if err != nil {
			return nil, fmt.Errorf("gitlab: failed to read root ca file %s: %w", rootCaFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("gitlab: no valid certificate found in root ca file %s", rootCaFile)
		}
		slog.Info("gitlab: loaded custom root ca", "file", rootCaFile)
		tlsConfig.RootCAs = pool
	}
	if rootCaDir != "" {
		entries, err := os.ReadDir(rootCaDir)
		if err != nil {
			return nil, fmt.Errorf("gitlab: failed to read root ca dir %s: %w", rootCaDir, err)
		}
		pool := tlsConfig.RootCAs
		if pool == nil {
			pool, err = x509.SystemCertPool()
			if err != nil {
				return nil, fmt.Errorf("gitlab: failed to load system cert pool: %w", err)
			}
		}
		loaded := 0
		for _, entry := range entries {
			if entry.IsDir() {
					continue
			}
			pem, err := os.ReadFile(fmt.Sprintf("%s/%s", rootCaDir, entry.Name()))
			if err != nil {
				slog.Warn("gitlab: failed to read ca file", "file", entry.Name(), "error", err)
				continue
			}
			if pool.AppendCertsFromPEM(pem) {
					loaded++
			} else {
				slog.Warn("gitlab: no valid certificate in ca file", "file", entry.Name())
			}
		}
		slog.Info("gitlab: loaded ca certificates from directory", "dir", rootCaDir, "count", loaded)
		if loaded == 0 {
			return nil, fmt.Errorf("gitlab: no valid ca certificate found in root ca dir %s", rootCaDir)
		}
		tlsConfig.RootCAs = pool
	}
	return tlsConfig, nil
}
