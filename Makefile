.PHONY: up down restart status logs setup pull-model gmail-auth build

# ─── First-time setup ───────────────────────────────────────

setup:
	@test -f .env && echo ".env already exists — skipping copy" || cp .env.example .env
	@if grep -q '^JWT_SECRET=$$' .env 2>/dev/null; then \
		SECRET=$$(openssl rand -base64 32); \
		sed -i '' "s|^JWT_SECRET=$$|JWT_SECRET=$$SECRET|" .env; \
		echo "Generated JWT_SECRET"; \
	else \
		echo "JWT_SECRET already set"; \
	fi
	@echo ""
	@echo "Setup complete. Next steps:"
	@echo "  1. Edit .env — fill in NGROK_AUTHTOKEN, NGROK_DOMAIN, and ADMIN_PASSWORD"
	@echo "  2. Run: make pull-model"
	@echo "  3. Run: make up"
	@echo "  4. Deploy frontend to Vercel (see README.md)"
	@echo "  5. Add your Vercel URL to VERCEL_URL in .env, then: make restart"

pull-model:
	@echo "Pulling Ollama model (llama3.2:3b)..."
	@ollama pull llama3.2:3b
	@echo "Model ready."

# ─── Gmail OAuth ────────────────────────────────────────────

gmail-auth:
	@echo "Starting Gmail OAuth flow..."
	@cd backend && go run cmd/gmailauth/main.go credentials.json token.json
	@echo ""
	@echo "Done. Restart the backend to enable Gmail polling: make restart"

# ─── Service management ────────────────────────────────────

up:
	@echo "Starting Ollama..."
	@pgrep -x ollama > /dev/null || (ollama serve &>/dev/null & sleep 2 && echo "Ollama started")
	@pgrep -x ollama > /dev/null && echo "Ollama running" || echo "Ollama failed to start"
	@echo "Starting Docker services..."
	@docker compose up -d
	@echo "All services up."

down:
	@echo "Stopping Docker services..."
	@docker compose down
	@echo "Stopping Ollama..."
	@pkill ollama 2>/dev/null || true
	@echo "All services stopped."

restart: down up

build:
	@echo "Rebuilding backend..."
	@docker compose up -d --build api
	@echo "Backend rebuilt."

status:
	@echo "=== Ollama ==="
	@pgrep -x ollama > /dev/null && echo "Running (PID $$(pgrep -x ollama))" || echo "Not running"
	@echo ""
	@echo "=== Docker ==="
	@docker compose ps

logs:
	@docker compose logs -f --tail=50
