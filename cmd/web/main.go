package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"uuid"

	"github.com/mwildt/git-workspace/pkg/gitlab"
)

type Error string

func (e Error) Error() string { return string(e) }

const NotFound = Error("Not Found")

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type (
	Project struct {
		Id     uuid.UUID `json:"id"`
		Path   string    `json:"path,omitempty"`
		GitUrl string    `json:"git_url,omitempty"`
		Branch string    `json:"branch,omitempty"`
	}

	Workspace struct {
		projects []Project
		baseDir  string
		mu       sync.RWMutex
	}

	Git struct {
		maxRetries    int
		retryDelay   time.Duration
		cloneTimeout time.Duration
	}

	Server struct {
		workspace    *Workspace
		gitlabClient *gitlab.Client
	}

	Config struct {
		Port         string
		BaseDir      string
		GitLabURL    string
		GitToken     string
		LogLevel     string
		CloneTimeout time.Duration
		MaxRetries   int
		RetryDelay   time.Duration
	}
)

func (s *Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	if projects, err := s.gitlabClient.ListProjects(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(projects); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(s.workspace.projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) GetRepoBranches(w http.ResponseWriter, request *http.Request) {
	projectid := request.PathValue("projectid")
	if projects, err := s.gitlabClient.ListBranches(projectid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(projects); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) PostWorkspaceProject(writer http.ResponseWriter, request *http.Request) {
	var project struct {
		Path    string `json:"path"`
		Project struct {
			GitUrl string `json:"http_url_to_repo"`
		} `json:"project"`
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	}

	err := json.NewDecoder(request.Body).Decode(&project)
	if err != nil {
		http.Error(writer, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.workspace.AddProject(project.Path, project.Project.GitUrl, project.Branch.Name)
	writer.WriteHeader(http.StatusCreated)
}

func (s *Server) DeleteWorkspaceProject(writer http.ResponseWriter, request *http.Request) {
	projectId := request.PathValue("projectId")
	if err := s.workspace.DeleteProjectById(projectId); err != nil && err != NotFound {
		slog.Error("failed to delete project", "projectId", projectId, "error", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusAccepted)
}

type InitResponse struct {
	Success bool     `json:"success"`
	Errors  []ErrorInfo `json:"errors,omitempty"`
}

type ErrorInfo struct {
	ProjectId string `json:"projectId"`
	ProjectPath string `json:"projectPath"`
	Message string `json:"message"`
}

func (s *Server) PostWorkspaceInit(writer http.ResponseWriter, request *http.Request) {
	git := &Git{
		maxRetries:    3,
		retryDelay:   5 * time.Second,
		cloneTimeout: 30 * time.Second,
	}
	errors := s.workspace.Initialize(git)
	
	response := InitResponse{
		Success: len(errors) == 0,
		Errors:  errors,
	}
	
	writer.Header().Set("Content-Type", "application/json")
	if response.Success {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusPartialContent)
	}
	
	if err := json.NewEncoder(writer).Encode(response); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (g *Git) Clone(url string, path string, branch string) error {
	slog.Info("cloning", "url", url, "to", path, "branch", branch)
	
	var lastErr error
	for i := 0; i <= g.maxRetries; i++ {
		if i > 0 {
			slog.Info("retrying clone", "attempt", i+1, "url", url)
			time.Sleep(g.retryDelay)
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), g.cloneTimeout)
		defer cancel()
		
		cmd := exec.CommandContext(ctx, "git", "clone", "--branch", branch, "--depth", "1", url, path)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		
		if err == nil {
			return nil
		}
		
		lastErr = err
		slog.Warn("clone attempt failed", "attempt", i+1, "error", err, "output", string(output))
		
		if ctx.Err() == context.DeadlineExceeded {
			break
		}
	}
	
	outputStr := strings.TrimSpace(string([]byte{}))
	if lastErr != nil {
		if outputStr != "" {
			return fmt.Errorf("git clone failed after %d attempts: %s (output: %s)", g.maxRetries+1, lastErr.Error(), outputStr)
		}
		return fmt.Errorf("git clone failed after %d attempts: %s", g.maxRetries+1, lastErr.Error())
	}
	return nil
}

func NewWorkspace(baseDir string) *Workspace {
	return &Workspace{
		projects: make([]Project, 0),
		baseDir:  baseDir,
	}
}

func (w *Workspace) AddProject(path string, gitUrl string, branch string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.projects = append(w.projects, Project{Path: path, GitUrl: gitUrl, Branch: branch, Id: uuid.New()})
}

func (w *Workspace) Initialize(git *Git) []ErrorInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	var errors []ErrorInfo
	for _, project := range w.projects {
		projectPath := path.Join(w.baseDir, project.Path)
		if err := os.MkdirAll(projectPath, 0755); err != nil {
			errors = append(errors, ErrorInfo{
				ProjectId:   project.Id.String(),
				ProjectPath: project.Path,
				Message:    fmt.Sprintf("failed to create directory: %s", err.Error()),
			})
			continue
		}
		if err := git.Clone(project.GitUrl, projectPath, project.Branch); err != nil {
			errors = append(errors, ErrorInfo{
				ProjectId:   project.Id.String(),
				ProjectPath: project.Path,
				Message:    err.Error(),
			})
		}
	}
	return errors
}

func (w *Workspace) DeleteProjectById(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	for i, project := range w.projects {
		if project.Id.String() == id {
			w.projects = append(w.projects[:i], w.projects[i+1:]...)
			return nil
		}
	}
	return NotFound
}

func loadConfig() Config {
	config := Config{
		Port:         "8080",
		BaseDir:      path.Join(os.Getenv("HOME"), ".git-workspace"),
		GitLabURL:    "https://gitlab.com",
		GitToken:     os.Getenv("GIT_TOKEN"),
		LogLevel:     "info",
		CloneTimeout: 30 * time.Second,
		MaxRetries:   3,
		RetryDelay:   5 * time.Second,
	}

	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}
	if baseDir := os.Getenv("WORKSPACE_BASE_DIR"); baseDir != "" {
		config.BaseDir = baseDir
	}
	if gitLabURL := os.Getenv("GITLAB_URL"); gitLabURL != "" {
		config.GitLabURL = gitLabURL
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}
	if cloneTimeout := os.Getenv("CLONE_TIMEOUT_SECONDS"); cloneTimeout != "" {
		if seconds, err := strconv.Atoi(cloneTimeout); err == nil && seconds > 0 {
			config.CloneTimeout = time.Duration(seconds) * time.Second
		}
	}
	if maxRetries := os.Getenv("MAX_RETRIES"); maxRetries != "" {
		if retries, err := strconv.Atoi(maxRetries); err == nil && retries >= 0 {
			config.MaxRetries = retries
		}
	}
	if retryDelay := os.Getenv("RETRY_DELAY_SECONDS"); retryDelay != "" {
		if seconds, err := strconv.Atoi(retryDelay); err == nil && seconds > 0 {
			config.RetryDelay = time.Duration(seconds) * time.Second
		}
	}

	return config
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func setupHandlers(server *Server, staticDir string) http.Handler {
	handler := http.NewServeMux()

	handler.HandleFunc("GET /api/gitlab/project/{projectid}/branches", server.GetRepoBranches)
	handler.HandleFunc("GET /api/gitlab/project", server.GetProjects)
	handler.HandleFunc("GET /api/workspace", server.GetWorkspace)
	handler.HandleFunc("POST /api/workspace/init", server.PostWorkspaceInit)
	handler.HandleFunc("POST /api/workspace/project", server.PostWorkspaceProject)
	handler.HandleFunc("DELETE /api/workspace/project/{projectId}", server.DeleteWorkspaceProject)

	handler.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	handler.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	handler.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, struct{ Message string }{Message: "Not Found"})
	})

	handler.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(staticDir, "index.html"))
	})

	return handler
}

func main() {
	config := loadConfig()
	logger := setupLogger(config.LogLevel)
	slog.SetDefault(logger)

	if config.GitToken == "" {
		logger.Error("GIT_TOKEN environment variable is required")
		os.Exit(1)
	}

	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		logger.Error("failed to create base directory", "path", config.BaseDir, "error", err)
		os.Exit(1)
	}

	server := Server{
		workspace:    NewWorkspace(config.BaseDir),
		gitlabClient: gitlab.NewClient(config.GitToken, config.GitLabURL),
	}

	staticDir := "./static"
	if staticDirEnv := os.Getenv("STATIC_DIR"); staticDirEnv != "" {
		staticDir = staticDirEnv
	}

	handler := setupHandlers(&server, staticDir)

	srv := &http.Server{
		Addr:              ":" + config.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	logger.Info("starting server", "port", config.Port, "baseDir", config.BaseDir, "gitLabURL", config.GitLabURL)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	
	<-quit
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
