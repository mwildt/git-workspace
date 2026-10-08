package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
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
	}

	Git struct{}

	Server struct {
		workspace    *Workspace
		gitlabClient *gitlab.Client
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
	if err := s.workspace.DeleteProjectById(projectId); err == NotFound {

	} else if err != nil {

	}
	writer.WriteHeader(http.StatusAccepted)

}

func (s *Server) PostWorkspaceInit(writer http.ResponseWriter, request *http.Request) {
	err := s.workspace.Initialize(&Git{})
	if err != nil {
		slog.Default().Error("Error initializing workspace", "error", err)
	}
}

func (g *Git) Clone(url string, path string, branch string) error {
	slog.Default().Info("cloning", url, "to", path, "branch", branch)
	cmd := exec.Command("git", "clone", "--branch", branch, url, path)
	cmd.Env = append(os.Environ())
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func NewWorkspace() *Workspace {
	return &Workspace{
		projects: make([]Project, 0),
		baseDir:  path.Join(os.Getenv("HOME"), ".git-workspace"),
	}
}

func (w *Workspace) AddProject(path string, gitUrl string, branch string) {
	w.projects = append(w.projects, Project{Path: path, GitUrl: gitUrl, Branch: branch, Id: uuid.New()})
}

func (w *Workspace) Initialize(git *Git) error {
	for _, project := range w.projects {
		path := path.Join(w.baseDir, project.Path)
		if err := git.Clone(project.GitUrl, path, project.Branch); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workspace) DeleteProjectById(id string) error {
	for i, project := range w.projects {
		if project.Id.String() == id {
			w.projects = append(w.projects[:i], w.projects[i+1:]...)
			return nil
		}
	}
	return NotFound
}

func main() {

	server := Server{
		workspace:    NewWorkspace(),
		gitlabClient: gitlab.NewClient(os.Getenv("GIT_TOKEN"), "https://gitlab.com"),
	}

	handler := http.NewServeMux()

	handler.HandleFunc("GET /api/gitlab/project/{projectid}/branches", server.GetRepoBranches)
	handler.HandleFunc("GET /api/gitlab/project", server.GetProjects)
	handler.HandleFunc("GET /api/workspace", server.GetWorkspace)
	handler.HandleFunc("POST /api/workspace/init", server.PostWorkspaceInit)
	handler.HandleFunc("POST /api/workspace/project", server.PostWorkspaceProject)
	handler.HandleFunc("DELETE /api/workspace/project/{projectId}", server.DeleteWorkspaceProject)

	handler.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	handler.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, struct{ Message string }{Message: "Not Found"})
	})

	handler.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	slog.Default().Info("starting server")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
