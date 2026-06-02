.PHONY: install dev web api speech

install:
	cd web && npm install
	cd services/api && go mod download
	cd services/speech-engine && pip install -r requirements.txt

dev:
	$(MAKE) -j3 speech api web

web:
	cd web && npm run dev

api:
	cd services/api && go run cmd/servers/session/main.go

speech:
	cd services/speech-engine && . .venv/bin/activate && python main.py -t webrtc