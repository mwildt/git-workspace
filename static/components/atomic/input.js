import { LitElement, html, css } from '../../lit.js';

customElements.define('gw-input', class extends LitElement {
    static properties = {
        value: { type: String, reflect: true },
        type: { type: String, reflect: true },
        placeholder: { type: String, reflect: true },
        disabled: { type: Boolean, reflect: true }
    };

    static styles = css`
        :host {
            display: block;
            margin: 8px 0;
        }

        input {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.3s ease, box-shadow 0.3s ease;
            box-sizing: border-box;
        }

        input:focus {
            outline: none;
            border-color: #4CAF50;
            box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
        }

        input:disabled {
            background-color: #f5f5f5;
            cursor: not-allowed;
        }

        label {
            display: block;
            margin-bottom: 4px;
            font-size: 14px;
            color: #555;
            font-weight: 500;
        }
    `;

    constructor() {
        super();
        this.type = 'text';
        this.value = '';
        this.placeholder = '';
        this.disabled = false;
    }

    handleInput(e) {
        this.value = e.target.value;
        this.dispatchEvent(new CustomEvent('input', { 
            detail: { value: this.value },
            bubbles: true,
            composed: true
        }));
    }

    render() {
        return html`
            <label><slot name="label"></slot></label>
            <input 
                .value=${this.value}
                type=${this.type}
                placeholder=${this.placeholder}
                ?disabled=${this.disabled}
                @input=${this.handleInput}
            >
        `;
    }
});