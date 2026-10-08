import { LitElement, html, css } from '../lit.js';
import './atomic/button.js';
import './atomic/error-message.js';

customElements.define('gw-workspace', class extends LitElement {
    static properties = {
        projects: {},
        errors: {},
        initializing: { type: Boolean }
    };

    static styles = css`
        :host {
            display: block;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: white;
            border-radius: 12px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            padding: 24px;
        }

        h4 {
            color: #2E7D32;
            margin: 0 0 20px 0;
            font-size: 24px;
            font-weight: 600;
            border-bottom: 3px solid #4CAF50;
            padding-bottom: 12px;
        }

        .project-list {
            margin: 20px 0;
        }

        .project-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 16px;
            margin: 8px 0;
            background-color: #f9f9f9;
            border-radius: 8px;
            border-left: 4px solid #4CAF50;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }

        .project-item:hover {
            transform: translateX(4px);
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }

        .project-info {
            flex: 1;
        }

        .project-info h5 {
            color: #2E7D32;
            margin: 0 0 4px 0;
            font-size: 16px;
            font-weight: 600;
        }

        .project-info span {
            color: #666;
            font-size: 14px;
            margin-right: 8px;
        }

        .project-meta {
            display: flex;
            gap: 16px;
            font-size: 13px;
        }

        .remove-btn {
            color: #f44336;
            cursor: pointer;
            font-size: 12px;
            text-decoration: underline;
            transition: color 0.2s ease;
        }

        .remove-btn:hover {
            color: #d32f2f;
        }

        .actions {
            display: flex;
            justify-content: flex-end;
            margin-top: 24px;
        }

        .empty-state {
            text-align: center;
            padding: 40px;
            color: #999;
            font-style: italic;
        }

        .status-message {
            padding: 12px 16px;
            border-radius: 8px;
            margin: 16px 0;
            font-size: 14px;
        }

        .success {
            background-color: #e8f5e8;
            color: #2E7D32;
            border-left: 4px solid #4CAF50;
        }

        .loading {
            background-color: #e3f2fd;
            color: #1976D2;
            border-left: 4px solid #2196F3;
        }

        .errors-container {
            margin: 16px 0;
        }

        .error-summary {
            background-color: #ffebee;
            color: #c62828;
            padding: 12px 16px;
            border-radius: 8px;
            margin-bottom: 12px;
            font-weight: 600;
            border-left: 4px solid #f44336;
        }
    `;

    constructor() {
        super();
        this.projects = [];
        this.errors = [];
        this.initializing = false;

        this.addEventListener('gw-add-project::success', async () => {
            await this.reload();
        });
    }

    async connectedCallback() {
        super.connectedCallback();
        await this.reload();
    }

    render() {
        return html`
            <h4>Projekte</h4>
            
            <gw-add-project></gw-add-project>
            
            ${this.initializing
                ? html`<div class="status-message loading">Initialisiere Projekte...</div>`
                : this.errors && this.errors.length > 0
                    ? html`
                        <div class="errors-container">
                            <div class="error-summary">
                                ${this.errors.length} Fehler beim Initialisieren aufgetreten
                            </div>
                            ${this.errors.map(error => html`<gw-error-message .error=${error}></gw-error-message>`)}
                        </div>
                    ` : undefined }
                ${this.projects.length > 0
                        ? html`
                            <div class="project-list">
                                ${this.projects.map(project => html`
                                    <div class="project-item">
                                        <div class="project-info">
                                            <h5>${project.path}</h5>
                                            <div class="project-meta">
                                                <span>${project.git_url}</span>
                                                <span>::</span>
                                                <span>${project.branch}</span>
                                            </div>
                                        </div>
                                        <span 
                                            class="remove-btn"
                                            @click=${() => this.removeProject(project)}
                                        >Entfernen</span>
                                    </div>
                                `)}
                            </div>
                            <div class="actions">
                                <gw-button @click=${() => this.initialize()}>Initialize</gw-button>
                            </div>
                        `
                        : html`<div class="empty-state">Keine Projekte vorhanden. Fügen Sie ein Projekt hinzu.</div>`}
        `;
    }

    async removeProject(project) {
        const delResponse = await fetch(`/api/workspace/project/${project.id}`, {
            method: "DELETE"
        });
        if (!delResponse.ok) {
            throw new Error(`HTTP ${delResponse.status}`);
        }
        await this.reload();
    }

    async reload() {
        const response = await fetch('/api/workspace');
        this.projects = await response.json();
    }

    async initialize() {
        this.initializing = true;
        this.errors = [];

        try {
            const response = await fetch('/api/workspace/init', {
                method: 'POST'
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                this.errors = data.errors || [];
            }
        } catch (error) {
            this.errors = [{ message: 'Netzwerkfehler: ' + error.message }];
        } finally {
            this.initializing = false;
        }
    }
});