import { LitElement, html, css } from '../../lit.js';

customElements.define('gw-error-message', class extends LitElement {
    static properties = {
        error: { type: Object }
    };

    static styles = css`
        :host {
            display: block;
            margin: 8px 0;
        }

        .error-message {
            background-color: #ffebee;
            color: #c62828;
            padding: 12px 16px;
            border-radius: 8px;
            border-left: 4px solid #f44336;
            font-size: 14px;
            margin: 8px 0;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .error-icon {
            font-size: 18px;
        }

        .error-content {
            flex: 1;
        }

        .error-project {
            font-weight: 600;
            margin-bottom: 4px;
        }

        .error-text {
            color: #d32f2f;
        }
    `;

    render() {
        if (!this.error) return null;

        return html`
            <div class="error-message">
                <span class="error-icon">⚠️</span>
                <div class="error-content">
                    <div class="error-project">${this.error.projectPath || this.error.projectId}</div>
                    <div class="error-text">${this.error.message}</div>
                </div>
            </div>
        `;
    }
});