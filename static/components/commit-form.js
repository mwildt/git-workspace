import { LitElement, html, css } from '../lit.js';
import './atomic/button.js';
import './atomic/input.js';
import './atomic/error-message.js';

customElements.define('gw-commit-form', class extends LitElement {
    static properties = {
        project: { },
        committing: { },
        commitMessage: { },
        commitBranch: {},
        commitError: { }
    };

    static styles = css`
        :host {
            display: block;
        }

        .commit-form {
            display: flex;
            flex-direction: column;
            gap: 8px;
            padding: 12px;
            margin: 8px 0;
            background-color: #f9f9f9;
            border-radius: 8px;
            border-left: 4px solid #4CAF50;
        }

        .commit-form-actions {
            display: flex;
            justify-content: flex-end;
            gap: 8px;
            align-items: center;
        }

        .commit-error {
            color: #c62828;
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
    `;

    constructor() {
        super();
        this.project = null;
        this.committing = false;
        this.commitMessage = '';
        this.commitBranch = '';
        this.commitError = null;
    }

    render() {
        if (!this.project) return undefined;

        return html`
            <div class="commit-form">
                <gw-input
                    placeholder="Commit Message" .value=${this.commitMessage}
                    @change=${e => this.commitMessage = e.detail.value}
                ></gw-input>
                <gw-input
                    placeholder="Zielbranch (optional, Default: HEAD)"  .value=${this.commitBranch}
                    @change=${e => this.commitBranch = e.detail.value}
                ></gw-input>
                ${this.commitError
                    ? html`<span class="commit-error">${this.commitError}</span>`
                    : undefined}
                <div class="commit-form-actions">
                    <gw-button
                        .disabled=${this.committing || !this.commitMessage}
                        @click=${() => this.dispatchCommit()}
                    >${this.committing ? 'Commite...' : 'Commit & Push'}</gw-button>
                    <span
                        class="remove-btn"
                        @click=${() => this.dispatchCancel()}
                    >Abbrechen</span>
                </div>
            </div>
        `;
    }

    dispatchCommit() {
        this.dispatchEvent(new CustomEvent('gw-commit-form::commit', {
            detail: {
                message: this.commitMessage,
                branch: this.commitBranch
            },
            bubbles: true,
            composed: true
        }));
    }

    dispatchCancel() {
        this.dispatchEvent(new CustomEvent('gw-commit-form::cancel', {
            bubbles: true,
            composed: true
        }));
    }

    reset() {
        this.commitMessage = '';
        this.commitBranch = '';
        this.commitError = null;
        this.committing = false;
    }
});
