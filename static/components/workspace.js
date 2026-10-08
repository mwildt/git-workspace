import { LitElement, html } from '../lit.js';

customElements.define('gw-workspace', class extends LitElement {
    static properties = {
        projects: {}
    }
    constructor() {
        super();
        this.projects = [];

        this.addEventListener('gw-add-project::success', async () => {
            await this.reload()
        })
    }

    async connectedCallback() {


        super.connectedCallback();

        await this.reload()
    }

    render() {
        return html`
            <h4>Projekte</h4>
            <gw-add-project></gw-add-project>
            ${this.projects.map(project => html`
                <div>
                    <h5>${project.path}</h5>
                    <span>${project.git_url}</span> :: <span>${project.branch}</span>
                    <small @click=${() => this.removeProject(project)}>Entfernen</small>
                </div>
            `)}

            <button @click=${() => this.initialize()}>Initialize</button>
            
      </ul>
    `;
    }

    async removeProject(project) {
        const delResponse = await fetch(`/api/workspace/project/${project.id}`, {
            method: "DELETE"
        });
        if (!delResponse.ok) {
            throw new Error(`HTTP ${delResponse.status}`);
        }
       await this.reload()
    }

    async reload() {
        const response = await fetch('/api/workspace');
        this.projects = await response.json();
    }

    async initialize() {
        const response = await fetch('/api/workspace/init', {
            method: 'POST'
        });
    }
})
