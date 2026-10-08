import { LitElement, html, css } from '../lit.js';

customElements.define('gw-add-project',  class extends LitElement {

    static properties = {
        model : {},
        projects : {},
        branches : {}
    }


    static styles = css`
        .selectable { padding: 6px 10px; cursor: pointer; }
        .selected { background: #d0e4ff; font-weight: 600; }
`;



    constructor() {
        super();
    }

    async connectedCallback() {
        super.connectedCallback();
    }

    render() {
        return html`
            ${this.model !== undefined
                    ? html`

                        
                        <h4>Project Wählen</h4>
                        ${this.projects && this.projects.map(project => html`
                            
                            <div class="selectable ${project == this.model.project ? "selected" : ""}"  @click=${() => this.selectProject(project)}>
                                <span>${project.name}</span>
                                <span>${project.web_url}</span>
                            </div>
                        `)}

                        ${this.model.project !== undefined
                            ? html`
                                <h4>Pfad Wählen</h4>
                                <span>Hier wird das Projekt nachher ausgecheckt</span>
                                <input type="text"  .value=${this.model ? this.model.path : undefined} @input=${e => this.model = {...this.model, path: e.target.value}} />


                                <h4>Branch Wählen</h4>
                                ${this.branches && this.branches.map(branch => html`
                                    <div class="selectable ${branch == this.model.branch ? "selected" : ""}" @click=${() => this.selectBranch(branch)}>
                                        <span>${branch.name}</span>
                                    </div>
                                `)}`
                            : undefined}

                        <pre >${JSON.stringify(this.model)}</pre>

                        <button @click=${() => this.submit()}>Hinzufügen</button>
                        <button @click=${() => this.model = undefined}>Abbrechen</button>
                    `
                    : html`<button @click=${() => this.start()}>Hinzufügen</button>`
        } 
         
    `;
    }

    async selectBranch(branch) {
        this.model = {...this.model, branch, }
    }

    async selectProject(project) {
        this.model = {...this.model, project, path: project.path_with_namespace }
        this.branches = undefined
        const response = await fetch(`/api/gitlab/project/${project.id}/branches`);
        this.branches = await response.json();
    }
    async start() {
        this.model = {}
        const response = await fetch('/api/gitlab/project');
        this.projects = await response.json();
    }

    async submit() {
        const response = await fetch(`/api/workspace/project`, {
            method: "POST",
            body: JSON.stringify(this.model),
        });
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        this.dispatchEvent(new CustomEvent('gw-add-project::success', {
            detail: this.model,
            bubbles: true,
            composed: true,
        }));

        this.model = undefined

    }
})
