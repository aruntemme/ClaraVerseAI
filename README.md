<div align="center">

<img src="docs/images/image-banner.png" alt="ClaraVerse - Your Private AI Workspace" width="800" />

### **Your Private AI Workspace**

*Built by the community, for the community. Private AI that respects your freedom.*

<p>

[![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/claraverse-space/ClaraVerseAI?style=social)](https://github.com/claraverse-space/ClaraVerseAI/stargazers)
[![Docker Pulls](https://img.shields.io/docker/pulls/claraverseoss/claraverse?color=blue)](https://hub.docker.com/r/claraverseoss/claraverse)
[![Discord](https://img.shields.io/badge/Discord-Join%20Us-7289da?logo=discord&logoColor=white)](https://discord.com/invite/j633fsrAne)

[Website](https://claraverse.space) · [Documentation](#documentation) · [Quick Start](#quick-start) · [Community](#community) · [Contributing](#contributing)

</div>

---

## What is ClaraVerse?

ClaraVerse is a private AI workspace that combines chat, image generation, and visual workflows in one app. Use OpenAI, Claude, Gemini, or local models — with browser-local storage that keeps your conversations private. Even server admins can't read your chats.

**50,000+ downloads** — the only AI platform where conversations never touch the server, even when self-hosted.

<img src="docs/images/clara-chat-demo.gif" alt="ClaraVerse Demo" width="800" />

### Try It Out

| Option | Description |
|--------|-------------|
| [**Cloud**](https://claraverse.app) | Free hosted version with TEE-secured inference — no setup required |
| [**Desktop**](https://github.com/claraverse-space/ClaraVerse) | Standalone Electron app for Windows, macOS, Linux (3.7k+ stars) |
| [**Self-Hosted**](#quick-start) | Docker deployment (this repo) — full control on your infrastructure |

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

---

## Quick Start

**Install CLI:**
```bash
curl -fsSL https://get.claraverse.app | bash
```

**Start ClaraVerse:**
```bash
claraverse init
```

Open **http://localhost** → Register → Add AI provider → Start chatting!

<details>
<summary><b>Other options</b></summary>

**Docker (no CLI):**
```bash
docker run -d -p 80:80 -p 3001:3001 -v claraverse-data:/data claraverseoss/claraverse:latest
```

**Clone & Run:**
```bash
git clone https://github.com/claraverse-space/ClaraVerseAI.git && cd ClaraVerseAI && ./quickstart.sh
```

</details>

<details>
<summary><b>Advanced Setup & Troubleshooting</b></summary>

### Prerequisites
- Docker & Docker Compose installed
- 4GB RAM minimum (8GB recommended)

### Manual Installation

```bash
# 1. Clone
git clone https://github.com/claraverse-space/ClaraVerseAI.git
cd ClaraVerseAI

# 2. Configure
cp .env.default .env

# 3. Start
docker compose up -d

# 4. Verify
docker compose ps
```

### Default Admin
```
Email: admin@localhost
Password: admin
```
**Change the password on first login.**

### Troubleshooting

```bash
# Run diagnostics
./diagnose.sh     # Linux/Mac
diagnose.bat      # Windows

# View logs
docker compose logs -f backend

# Restart
docker compose restart

# Fresh start
docker compose down -v && docker compose up -d
```

</details>

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
| [Architecture Guide](docs/ARCHITECTURE.md) | System design and components |
| [API Reference](docs/API_REFERENCE.md) | REST and WebSocket API |
| [Developer Guide](docs/DEVELOPER_GUIDE.md) | Local development setup |
| [Security Guide](docs/FINAL_SECURITY_INSPECTION.md) | Security features |
| [Admin Guide](docs/ADMIN_GUIDE.md) | System administration |
| [Quick Reference](docs/QUICK_REFERENCE.md) | Common commands |

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

Build your own private AI workspace with ClaraVerseAI 🌸

</div>
