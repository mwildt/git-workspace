import { LitElement, html, css } from '../../lit.js';

customElements.define('gw-selectable', class extends LitElement {
    static properties = {
        selected: { type: Boolean, reflect: true }
    };

    static styles = css`
        :host {
            display: block;
            margin: 4px 0;
        }

        .selectable-item {
            padding: 12px 16px;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.2s ease;
            border: 2px solid transparent;
            background-color: white;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .selectable-item:hover {
            background-color: #f5f5f5;
            border-color: #e0e0e0;
        }

        .selectable-item.selected {
            background-color: #e8f5e8;
            border-color: #4CAF50;
            font-weight: 600;
        }

        .selectable-item span {
            color: #333;
        }

        .selectable-item.selected span:first-child {
            color: #2E7D32;
        }

        .url {
            font-size: 12px;
            color: #666;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            max-width: 200px;
        }
    `;

    constructor() {
        super();
        this.selected = false;
    }

    handleClick() {
        this.dispatchEvent(new CustomEvent('click', {
            bubbles: true,
            composed: true
        }));
    }

    render() {
        return html`
            <div 
                class="selectable-item ${this.selected ? 'selected' : ''}"
                @click=${this.handleClick}
            >
                <span><slot name="primary"></slot></span>
                <span class="url"><slot name="secondary"></slot></span>
            </div>
        `;
    }
});