# Git Workspace

Ein Web-Service zum Verwalten von Git-Workspaces mit Integration in GitLab.

## Features

- GitLab-Projekte auflisten und durchsuchen
- Workspace mit mehreren Git-Projekten verwalten
- Projekte klonen mit konfigurierbaren Branches
- Resiliente Git-Operationen mit Retries und Timeouts
- REST API für die Integration in andere Anwendungen

## Schnellstart

### Voraussetzungen

- Go 1.27+
- Git
- GitLab Personal Access Token

### Installation

```bash
# Klonen und Bauen
git clone <repository-url>
cd git-workspace
go build -o git-workspace ./cmd/web

# Ausführen
GIT_TOKEN=your-gitlab-token ./git-workspace
```

### Docker

```bash
# Image bauen
docker build -t git-workspace .

# Container starten
docker run -d \
  -e GIT_TOKEN=your-gitlab-token \
  -e GITLAB_URL=https://gitlab.com \
  -e PORT=8080 \
  -p 8080:8080 \
  git-workspace
```

### Eigene Root-CA (z. B. firmeninterne GitLab-Instanz)

Das Zertifikat kann per Volume-Mount in den Container eingebracht werden:

```bash
docker run -d \
  -e GIT_TOKEN=your-gitlab-token \
  -e GITLAB_URL=https://gitlab.example.com \
  -e ROOT_CA_FILE=/certs/root-ca.crt \
  -v /path/to/your/root-ca.crt:/certs/root-ca.crt:ro \
  -p 8080:8080 \
  git-workspace
```

Alternativ ein Verzeichnis mit mehreren CA-Zertifikaten (`ROOT_CA_DIR`) – dort werden alle PEM-Dateien geladen. Die CA gilt sowohl für die GitLab-API-Calls als auch für die Git-Operationen (clone/push).

## Konfiguration

### Environment-Variablen

| Variable | Beschreibung | Default |
|----------|--------------|---------|
| `PORT` | Server-Port | 8080 |
| `GIT_TOKEN` | GitLab Personal Access Token (erforderlich) | - |
| `ACCESS_TOKEN` | Access-Token für Login/Session-Schutz. Wenn gesetzt, sind alle API-Endpunkte geschützt | - |
| `GITLAB_URL` | GitLab-Instanz URL | https://gitlab.com |
| `WORKSPACE_BASE_DIR` | Basisverzeichnis für Workspaces | ~/.git-workspace |
| `STATIC_DIR` | Pfad zu statischen Dateien | ./static |
| `LOG_LEVEL` | Log-Level (debug, info, warn, error) | info |
| `CLONE_TIMEOUT_SECONDS` | Timeout für Git-Clone in Sekunden | 30 |
| `MAX_RETRIES` | Maximaler Wiederholungsversuche für Clone | 3 |
| `RETRY_DELAY_SECONDS` | Verzögerung zwischen Retries in Sekunden | 5 |
| `ROOT_CA_FILE` | Pfad zu einer PEM-Datei mit Root-CA, die den HTTPS-Calls gegen GitLab vertraut wird (auch für Git-Operationen) | - |
| `ROOT_CA_DIR` | Verzeichnis mit PEM-Dateien von Root-CAs (alle werden geladen) | - |

### Beispiel Konfiguration

```bash
export GIT_TOKEN="your-personal-access-token"
export GITLAB_URL="https://gitlab.example.com"
export PORT="8080"
export WORKSPACE_BASE_DIR="/opt/git-workspace"
export LOG_LEVEL="debug"
export CLONE_TIMEOUT_SECONDS="60"
export MAX_RETRIES="5"
export RETRY_DELAY_SECONDS="10"
```

## API Endpunkte

### Auth

| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| POST | `/api/auth/login` | Login mit `{"token": "..."}` – eröffnet eine Session (Cookie) |
| POST | `/api/auth/logout` | Session beenden |
| GET | `/api/auth/session` | Session-Status abfragen (200 = gültige Session, 401 = keine Session) |

### GitLab

| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/api/gitlab/project` | Liste aller GitLab-Projekte |
| GET | `/api/gitlab/project/{projectid}/branches` | Liste aller Branches eines Projekts |

### Workspace

| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/api/workspace` | Liste aller Projekte im Workspace |
| POST | `/api/workspace/project` | Projekt zum Workspace hinzufügen |
| DELETE | `/api/workspace/project/{projectId}` | Projekt aus Workspace entfernen |
| POST | `/api/workspace/init` | Alle Projekte im Workspace klonen |

### Health

| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/health` | Health-Check Endpunkt |

## Request/Response Beispiele

### Projekt hinzufügen

**Request:**
```bash
curl -X POST http://localhost:8080/api/workspace/project \
  -H "Content-Type: application/json" \
  -d '{
    "path": "mein-projekt",
    "project": {
      "http_url_to_repo": "https://gitlab.com/username/repo.git"
    },
    "branch": {
      "name": "main"
    }
  }'
```

**Response:**
```json
HTTP/1.1 201 Created
```

### Workspace initialisieren

**Request:**
```bash
curl -X POST http://localhost:8080/api/workspace/init
```

**Response (erfolgreich):**
```json
HTTP/1.1 200 OK
{
  "success": true,
  "errors": []
}
```

**Response (teilweise fehlgeschlagen):**
```json
HTTP/1.1 206 Partial Content
{
  "success": false,
  "errors": [
    {
      "projectId": "abc-123",
      "projectPath": "mein-projekt",
      "message": "git clone failed after 3 attempts: ..."
    }
  ]
}
```

## Entwicklung

### Projekt struktur

```
git-workspace/
├── cmd/
│   └── web/
│       └── main.go      # Hauptanwendung
├── pkg/
│   └── gitlab/
│       └── gitlab.go   # GitLab Client
├── static/             # Statische Dateien (Frontend)
├── go.mod
└── README.md
```

### Lokale Entwicklung

```bash
# Server starten
go run ./cmd/web

# Tests ausführen
go test ./...
```

## Lizenz

MIT License
