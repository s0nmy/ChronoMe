.PHONY: dev backend frontend b f

dev:
	$(MAKE) -j 2 backend frontend

backend:
	cd backend && go run ./cmd/server

b: backend

frontend:
	cd frontend && npm run dev

f: frontend
