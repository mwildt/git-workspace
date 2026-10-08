import { LitElement, html, css } from '../../lit.js';

customElements.define('gw-button', class extends LitElement {
    static properties = {
        type: { type: String, reflect: true },
        disabled: { type: Boolean, reflect: true }
    };

    static styles = css`
        :host {
            display: inline-block;
        }

        button {
            background-color: #4CAF50;
            color: white;
            border: none;
            padding: 10px 20px;
            text-align: center;
            text-decoration: none;
            display: inline-block;
            font-size: 14px;
            margin: 4px 2px;
            cursor: pointer;
            border-radius: 8px;
            transition: background-color 0.3s ease, transform 0.2s ease;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        button:hover:not(:disabled) {
            background-color: #45a049;
            transform: translateY(-1px);
        }

        button:active:not(:disabled) {
            transform: translateY(0);
        }

        button:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }

        button.secondary {
            background-color: #f0f0f0;
            color: #333;
        }

        button.secondary:hover:not(:disabled) {
            background-color: #e0e0e0;
        }

        button.danger {
            background-color: #f44336;
        }

        button.danger:hover:not(:disabled) {
            background-color: #d32f2f;
        }
    `;

    constructor() {
        super();
        this.type = 'primary';
        this.disabled = false;
    }

    render() {
        return html`
            <button 
                ?disabled=${this.disabled}
                class=${this.type}
            >
                <slot></slot>
            </button>
        `;
    }
});