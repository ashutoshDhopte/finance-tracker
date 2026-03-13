# Finance Tracker

A personal finance tracker that automatically imports transactions from bank emails (Gmail) and CSV files, categorizes them using a local LLM, and provides dashboards and reports.

## Architecture

```
Browser  -->  Vercel (Next.js frontend)
         -->  ngrok tunnel  -->  Local Docker (Go API + Postgres)
                                      |
                                      +--> Ollama (local LLM for categorization)
                                      +--> Gmail API (email polling)
```

- **Frontend**: Next.js hosted on Vercel (free tier)
- **Backend**: Go (Gin) API running locally in Docker
- **Database**: PostgreSQL 16 running locally in Docker
- **Tunnel**: ngrok exposes the local API to the internet (free tier)
- **LLM**: Ollama running on the host for transaction categorization
- **Email**: Gmail API for automatic bank email parsing (optional)

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Ollama](https://ollama.com/) — `brew install ollama`
- [Git](https://git-scm.com/)
- A free [ngrok account](https://ngrok.com/) (for tunneling)
- A free [Vercel account](https://vercel.com/) (for frontend hosting)
- [Go 1.25+](https://go.dev/) (only needed for Gmail OAuth setup)

## Quick Start

### 1. Clone and setup

```bash
git clone <repo-url>
cd finance-tracker
make setup
```

This creates `.env` from the template and generates a JWT secret.

### 2. Configure environment

Edit `.env` and fill in:

| Variable | Where to get it |
|----------|----------------|
| `ADMIN_PASSWORD` | Choose a strong password |
| `NGROK_AUTHTOKEN` | [ngrok dashboard](https://dashboard.ngrok.com/get-started/your-authtoken) |
| `NGROK_DOMAIN` | [ngrok domains](https://dashboard.ngrok.com/domains) — claim 1 free static domain |

### 3. Pull the LLM model

```bash
make pull-model
```

This downloads `llama3.2:3b` (~2GB) which is used for transaction categorization.

### 4. Start everything

```bash
make up
```

This starts Ollama on your host + Postgres, the API, and the ngrok tunnel in Docker.

Verify it's working:

```bash
make status
curl https://<your-ngrok-domain>/health
```

### 5. Deploy frontend to Vercel

1. Push this repo to GitHub
2. Go to [vercel.com](https://vercel.com/) — "New Project" — import your repo
3. Set the **root directory** to `frontend`
4. Add this environment variable in Vercel project settings:
   ```
   NEXT_PUBLIC_API_URL = https://<your-ngrok-domain>/api/v1
   ```
5. Deploy

After Vercel gives you a URL (e.g., `your-app.vercel.app`):

1. Set `VERCEL_URL` in your `.env` (without `https://` or trailing slash):
   ```
   VERCEL_URL=your-app.vercel.app
   ```
2. Restart the backend to apply the CORS change:
   ```bash
   make restart
   ```

### 6. Log in

Open your Vercel URL in a browser and log in with the credentials from `.env` (`ADMIN_USERNAME` / `ADMIN_PASSWORD`).

## Gmail Integration (Optional)

Gmail integration automatically polls your inbox for bank transaction emails and imports them. This requires a Google Cloud OAuth setup.

### First-time setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or use an existing one)
3. Enable the **Gmail API**: APIs & Services > Enable APIs > search "Gmail API" > Enable
4. Configure the **OAuth consent screen**:
   - User type: External
   - App name: Finance Tracker
   - Scopes: Add `https://www.googleapis.com/auth/gmail.readonly`
   - Test users: Add your Gmail address (and friends' emails if sharing)
5. Create **OAuth credentials**:
   - APIs & Services > Credentials > Create Credentials > OAuth Client ID
   - Application type: Desktop app
   - Download the JSON and save it as `backend/credentials.json`
6. Run the auth flow:
   ```bash
   make gmail-auth
   ```
   This opens a URL — log in with your Gmail, authorize, paste the code back.
7. Restart the backend:
   ```bash
   make restart
   ```

### Re-authorizing (token expired)

If your Google Cloud project is in **Testing mode** (not published), OAuth tokens expire every **7 days**. When this happens, Gmail polling will start failing with auth errors in the logs.

To re-authorize:

```bash
# 1. Stop the backend
make down

# 2. Delete the expired token
rm backend/token.json

# 3. Re-run the OAuth flow
make gmail-auth

# 4. Start everything back up
make up
```

To avoid the 7-day expiry, you can publish your Google Cloud app (requires Google verification since Gmail is a sensitive scope). For personal use, re-running `make gmail-auth` weekly is the simplest approach.

### Customizing the email query

The `GMAIL_QUERY` variable in `.env` controls which emails are scanned. Update it to match your bank's alert sender addresses:

```env
GMAIL_QUERY=from:(alerts@yourbank.com OR notifications@otherbank.com)
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make setup` | First-time setup — creates `.env`, generates JWT secret |
| `make up` | Start Ollama + all Docker services |
| `make down` | Stop everything |
| `make restart` | Stop then start |
| `make build` | Rebuild the backend (after code changes) |
| `make status` | Show running status of all services |
| `make logs` | Tail Docker logs |
| `make pull-model` | Download the Ollama LLM model |
| `make gmail-auth` | Run Gmail OAuth flow to generate/refresh token |

## Importing Transactions

### From CSV

1. Go to the Transactions page in the UI
2. Click "Import CSV"
3. Select your bank account and upload a CSV file
4. Supported formats: Chase and PNC (auto-detected)

### From Gmail

1. Complete the Gmail Integration setup above
2. Go to the UI and click "Sync Gmail" on the dashboard, or the backend automatically polls based on `GMAIL_POLL_INTERVAL`

## Project Structure

```
finance-tracker/
├── backend/                  # Go API server
│   ├── cmd/
│   │   ├── server/           # Main server entrypoint
│   │   └── gmailauth/        # Gmail OAuth CLI tool
│   ├── internal/
│   │   ├── api/              # HTTP handlers, middleware, router
│   │   ├── config/           # Configuration loading
│   │   ├── db/migrations/    # SQL migrations
│   │   ├── models/           # Data models
│   │   └── services/         # Business logic (parser, gmail, dedup)
│   └── Dockerfile
├── frontend/                 # Next.js frontend
│   ├── src/
│   │   ├── app/              # Pages (dashboard, transactions, reports, etc.)
│   │   ├── components/       # Shared UI components
│   │   └── lib/              # API client, types, utilities
│   └── Dockerfile
├── docker-compose.yml        # Postgres + API + ngrok
├── Makefile                  # All management commands
├── .env.example              # Environment template
└── .gitignore
```
