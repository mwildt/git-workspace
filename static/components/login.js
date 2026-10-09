import { LitElement, html, css } from '../lit.js';
import './atomic/button.js';
import './atomic/input.js';

customElements.define('gw-login', class extends LitElement {
    static properties = {
        error: { type: String },
        loading: { type: Boolean },
    };

    static styles = css`
        :host {
            display: block;
            position: fixed;
            inset: 0;
            background-color: rgba(0, 0, 0, 0.5);
            z-index: 1000;
        }
        .card {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            width: min(400px, 90vw);
            padding: 32px;
            background-color: white;
            border-radius: 12px;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
        }
        h3 {
            color: #2E7D32;
            font-size: 20px;
            font-weight: 600;
            margin: 0 0 8px 0;
        }
        p {
            color: #666;
            font-size: 14px;
            margin-bottom: 20px;
        }
        gw-input {
            display: block;
            margin-bottom: 20px;
        }
        .error {
            background-color: #ffebee;
            color: #c62828;
            padding: 10px 14px;
            border-radius: 8px;
            font-size: 14px;
            margin-bottom: 16px;
            border-left: 4px solid #f44336;
        }
        .actions {
            display: flex;
            justify-content: flex-end;
        }
    `;

    constructor() {
        super();
        this.error = '';
        this.loading = false;
    }

    render() {
        return html`
            <div class="card">
                <h3>Anmeldung erforderlich</h3>
                <p>Bitte gib deinen Access-Token ein, um fortzufahren.</p>
                ${this.error ? html`<div class="error">${this.error}</div>` : undefined}
                <gw-input type="password" .value=${this.token || ''} @input=${(e) => this.token = e.target.value} @keydown=${(e) => e.key === 'Enter' && this.submit(e)}>
                    <span slot="label">Access Token</span>
                </gw-input>
                <div class="actions">
                    <gw-button ?disabled=${this.loading} @click=${this.submit}>
                        ${this.loading ? 'Anmelden...' : 'Anmelden'}
                    </gw-button>
                </div>
            </div>
        `;
    }

    async submit(event) {
        if (event) event.preventDefault();
        this.loading = true;
        this.error = '';
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token: this.token }),
            });
            if (!response.ok) {
                this.error = 'Anmeldung fehlgeschlagen. Bitte Token prüfen.';
                return;
            }
            this.dispatchEvent(new CustomEvent('gw-login::success', { bubbles: true, composed: true }));
        } catch (e) {
            this.error = 'Netzwerkfehler: ' + e.message;
        } finally {
            this.loading = false;
        }
    }
});
