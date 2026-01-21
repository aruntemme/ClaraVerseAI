<div align="center">

<img src="docs/images/image-banner.png" alt="ClaraVerse - Your Private AI Workspace" width="800" />

### Your Private AI Workspace

*One app replaces ChatGPT, Midjourney, and N8N. Local or cloud — your data stays yours.*

[![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/claraverse-space/ClaraVerseAI?style=social)](https://github.com/claraverse-space/ClaraVerseAI/stargazers)
[![GitHub Release](https://img.shields.io/github/v/release/claraverse-space/ClaraVerseAI)](https://github.com/claraverse-space/ClaraVerseAI/releases)
[![Discord](https://img.shields.io/discord/1332357374635888713?color=7289da&label=Discord&logo=discord&logoColor=white)](https://discord.com/invite/j633fsrAne)

[Website](https://claraverse.space) | [Documentation](backend/docs/) | [Discord](https://discord.com/invite/j633fsrAne)

</div>

---

## What is ClaraVerse?

ClaraVerse is a private AI workspace that combines chat, image generation, and visual workflows in one app. Use OpenAI, Claude, Gemini, or local models — with browser-local storage that keeps your conversations private. Even server admins can't read your chats.

**50,000+ downloads** — the only AI platform where conversations never touch the server, even when self-hosted.

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Browser-Local Storage** | Conversations stay in IndexedDB — zero-knowledge architecture |
| **Multi-Provider** | OpenAI, Anthropic, Google, Ollama, any OpenAI-compatible endpoint |
| **Visual Workflows** | Drag-and-drop builder — chat to design, visual editor to refine |
| **MCP Bridge** | Native Model Context Protocol support for seamless tool connections |
| **Interactive Prompts** | AI asks clarifying questions mid-conversation with typed forms |
| **Offline Mode** | Works completely air-gapped after initial load |
| **Code Execution** | Run Python/JS in sandboxed E2B (local Docker, no API key needed) |
| **Memory System** | Clara remembers context, auto-archives what's not used |
| **BYOK** | Bring your own API keys or use free local models |
| **Cross-Platform** | Web, Desktop (Windows, macOS, Linux), mobile-ready |

---

## Quick Start

### Prerequisites
- Docker & Docker Compose
- 4GB RAM minimum (8GB recommended)

### Installation

```bash
# Clone the repository
git clone https://github.com/claraverse-space/ClaraVerseAI.git
cd ClaraVerseAI

# Configure environment (generate secrets)
# ENCRYPTION_MASTER_KEY: openssl rand -hex 32
# JWT_SECRET: openssl rand -hex 64

# Start all services
docker compose up -d
```

### Access Points
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:3001

### Default Admin
```
Email: admin@localhost
Password: admin
```
**Change the password on first login.**

### What's Running
- Frontend (React + Vite)
- Backend (Go API)
- MongoDB, MySQL, Redis
- E2B (local code execution)
- SearXNG (self-hosted search)

---

## Development Setup

For contributors and local development:

```bash
# Prerequisites: Go 1.24+, Node.js 20+, Python 3.11+, tmux

# Install dependencies
make install

# Launch all services in tmux
./dev.sh
```

Opens a 4-pane terminal with Backend, Frontend, E2B Service, and Info panel.

---

## Tech Stack

- **Frontend**: React 19, TypeScript, Vite, Tailwind CSS 4, Zustand
- **Backend**: Go 1.24, Fiber, WebSocket streaming
- **Database**: MongoDB, MySQL, Redis
- **Services**: SearXNG (search), E2B Local Docker (code execution)
- **Auth**: Local JWT with Argon2id password hashing

---

## Documentation

| Resource | Description |
|----------|-------------|
| [Architecture Guide](backend/docs/ARCHITECTURE.md) | System design and components |
| [API Reference](backend/docs/API_REFERENCE.md) | REST and WebSocket API |
| [Developer Guide](backend/docs/DEVELOPER_GUIDE.md) | Local development setup |
| [Security Guide](backend/docs/FINAL_SECURITY_INSPECTION.md) | Security features |
| [Admin Guide](backend/docs/ADMIN_GUIDE.md) | System administration |
| [Quick Reference](backend/docs/QUICK_REFERENCE.md) | Common commands |

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make changes and test
4. Commit: `git commit -m 'Add amazing feature'`
5. Push and open a Pull Request

**Areas we need help:**
- Bug fixes ([open issues](https://github.com/claraverse-space/ClaraVerseAI/issues))
- Documentation improvements
- Translations
- New integrations and model providers

---

## Community

- [Discord](https://discord.com/invite/j633fsrAne) — Chat and support
- [Twitter/X](https://x.com/claraversehq) — Updates
- [GitHub Issues](https://github.com/claraverse-space/ClaraVerseAI/issues) — Bug reports
- [GitHub Discussions](https://github.com/claraverse-space/ClaraVerseAI/discussions) — Feature requests

---

## License

**AGPL-3.0** — Free to use, modify, and host commercially. Modifications must be open-sourced. See [LICENSE](LICENSE) for details.

---

<div align="center">

**Built with love by the ClaraVerse Community**

[Back to Top](#your-private-ai-workspace)

</div>
