import { LitElement, html, css } from '../lit.js';
import './atomic/header.js';
import './workspace.js';
import './login.js';

customElements.define('gw-app', class extends LitElement {
    static properties = {
        authenticated: { type: Boolean },
        checking: { type: Boolean },
    };

    static styles = css`
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .page-title {
            color: #2E7D32;
            margin-bottom: 8px;
            font-size: 24px;
            font-weight: 600;
        }
        .page-subtitle {
            color: #666;
            font-size: 16px;
            margin-bottom: 24px;
        }
    `;

    constructor() {
        super();
        this.authenticated = false;
        this.checking = true;
    }

    async connectedCallback() {
        super.connectedCallback();
        await this.checkSession();
        this.addEventListener('gw-login::success', async () => {
            this.authenticated = true;
            await this.reloadWorkspace();
        });
    }

    async checkSession() {
        this.checking = true;
        try {
            const response = await fetch('/api/auth/session');
            const data = await response.json();
            this.authenticated = !!data.authenticated;
        } catch (e) {
            this.authenticated = false;
        } finally {
            this.checking = false;
        }
    }

    async reloadWorkspace() {
        const workspace = this.renderRoot.querySelector('gw-workspace');
        if (workspace) {
            await workspace.reload();
        }
    }

    render() {
        if (this.checking) {
            return html`<gw-header></gw-header>`;
        }
        if (!this.authenticated) {
            return html`
                <gw-header></gw-header>
                <gw-login></gw-login>
            `;
        }
        return html`
            <gw-header></gw-header>
            <div class="container">
                <h1 class="page-title">Workspace Verwaltung</h1>
                <p class="page-subtitle">Durchsuche deine Git-Projekte und erstelle deinen Arbeitsbereich</p>
                <gw-workspace></gw-workspace>
            </div>
        `;
    }
});
