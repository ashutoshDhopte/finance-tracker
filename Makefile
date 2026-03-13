.PHONY: up down restart status logs

up:
	@echo "Starting Ollama..."
	@pgrep -x ollama > /dev/null || (/opt/homebrew/bin/ollama serve &>/dev/null & sleep 2 && echo "Ollama started")
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

status:
	@echo "=== Ollama ==="
	@pgrep -x ollama > /dev/null && echo "Running (PID $$(pgrep -x ollama))" || echo "Not running"
	@echo ""
	@echo "=== Docker ==="
	@docker compose ps

logs:
	@docker compose logs -f --tail=50
