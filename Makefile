.PHONY: dev web api speech

dev:
	$(MAKE) -j3 speech api web

web:
	cd web && npm run dev

api:
	cd services/api && go run cmd/servers/model-gateway/main.go

speech:
	cd services/speech-engine && . .venv/bin/activate && python main.py -t webrtc