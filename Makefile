.PHONY: install dev web api speech api-migrate-up

install:
	cd web && npm install
	cd services/api && go mod download
	cd services/speech-engine && pip install -r requirements.txt

dev:
	$(MAKE) -j3 speech api web

web:
	cd web && npm run dev

api:
	cd services/api && go run cmd/server/main.go

speech:
	cd services/speech-engine && . .venv/bin/activate && python main.py

api-migrate-up:
	cd services/api && make migrate-up