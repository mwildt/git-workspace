import { LitElement, html, css } from '../../lit.js';

customElements.define('gw-header', class extends LitElement {
    static styles = css`
        :host {
            display: block;
            background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%);
            color: white;
            padding: 24px 0;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            margin-bottom: 32px;
        }

        .header-content {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 24px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
            letter-spacing: -0.5px;
        }

        .subtitle {
            margin: 8px 0 0 0;
            font-size: 16px;
            opacity: 0.9;
            font-weight: 400;
        }

        .logo {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .logo-icon {
            width: 40px;
            height: 40px;
            background-color: rgba(255, 255, 255, 0.2);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
        }

        .user-info {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .user-avatar {
            width: 40px;
            height: 40px;
            background-color: rgba(255, 255, 255, 0.2);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    `;

    render() {
        return html`
            <header>
                <div class="header-content">
                    <div class="logo">
                        <div class="logo-icon">🌿</div>
                        <div>
                            <h1>Git Workspaces</h1>
                        </div>
                    </div>
                </div>
            </header>
        `;
    }
});