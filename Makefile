# ── Local development ─────────────────────────────────────────────────────────
dev:
	APP_ENV=dev go run main.go

stag:
	APP_ENV=stag go run main.go

prod:
	APP_ENV=prod go run main.go

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -trimpath -ldflags="-s -w" -o suberes_app ./main.go

# ── Tests ─────────────────────────────────────────────────────────────────────
test:
	go test ./... -race -cover

# ── Docker ───────────────────────────────────────────────────────────────────
IMAGE_TAG ?= latest

docker-build:
	docker build -t suberes_app:$(IMAGE_TAG) .

docker-staging:
	docker compose -f docker-compose.staging.yml up -d

docker-staging-down:
	docker compose -f docker-compose.staging.yml down

docker-production:
	docker compose -f docker-compose.production.yml up -d

docker-production-down:
	docker compose -f docker-compose.production.yml down

docker-logs:
	docker compose -f docker-compose.production.yml logs -f app

# ── SSL: cek masa berlaku sertifikat ─────────────────────────────────────────
# Usage: make ssl-check DOMAIN=api.suberes.com
DOMAIN ?= api.suberes.com
CERT_PATH ?= /etc/letsencrypt/live/$(DOMAIN)/fullchain.pem

ssl-check:
	@echo "==> Checking SSL certificate for $(DOMAIN)"
	@if docker run --rm \
	    -v /etc/letsencrypt:/etc/letsencrypt:ro \
	    alpine:3.21 sh -c \
	    "apk add -q openssl && \
	     if [ ! -f '$(CERT_PATH)' ]; then echo 'NOT FOUND'; exit 1; fi && \
	     EXPIRY=\$$(openssl x509 -noout -enddate -in '$(CERT_PATH)' | cut -d= -f2) && \
	     EXPIRY_EPOCH=\$$(date -d \"\$$EXPIRY\" +%s 2>/dev/null || date -j -f '%b %d %H:%M:%S %Y %Z' \"\$$EXPIRY\" +%s) && \
	     NOW_EPOCH=\$$(date +%s) && \
	     DAYS_LEFT=\$$(((\$$EXPIRY_EPOCH - \$$NOW_EPOCH) / 86400)) && \
	     echo \"Certificate expires: \$$EXPIRY\" && \
	     echo \"Days remaining    : \$$DAYS_LEFT\" && \
	     if [ \$$DAYS_LEFT -lt 14 ]; then echo 'WARNING: certificate expires in less than 14 days!'; exit 2; \
	     else echo 'Certificate is valid.'; fi"; then \
	  echo "OK"; \
	else \
	  STATUS=$$?; \
	  if [ $$STATUS -eq 1 ]; then echo "Certificate NOT FOUND — run: make ssl-init DOMAIN=$(DOMAIN)"; fi; \
	  if [ $$STATUS -eq 2 ]; then echo "ACTION NEEDED: renew soon — run: make ssl-renew DOMAIN=$(DOMAIN)"; fi; \
	fi

# ── SSL: issue certificate hanya jika belum ada atau hampir expired ───────────
# Aman dijalankan berkali-kali. Akan skip jika sertifikat masih berlaku > 30 hari.
# Usage: make ssl-init DOMAIN=api.suberes.com EMAIL=admin@suberes.com
EMAIL ?= admin@suberes.com

ssl-init:
	@echo "==> Checking existing certificate for $(DOMAIN)..."
	@CERT_EXISTS=0; DAYS_LEFT=0; \
	if docker run --rm -v /etc/letsencrypt:/etc/letsencrypt:ro alpine:3.21 sh -c \
	    "apk add -q openssl && [ -f '$(CERT_PATH)' ]" 2>/dev/null; then \
	  CERT_EXISTS=1; \
	  DAYS_LEFT=$$(docker run --rm -v /etc/letsencrypt:/etc/letsencrypt:ro alpine:3.21 sh -c \
	    "apk add -q openssl && \
	     EXPIRY=\$$(openssl x509 -noout -enddate -in '$(CERT_PATH)' | cut -d= -f2) && \
	     EXPIRY_EPOCH=\$$(date -d \"\$$EXPIRY\" +%s 2>/dev/null || date +%s) && \
	     echo \$$(((\$$EXPIRY_EPOCH - \$$(date +%s)) / 86400))"); \
	fi; \
	if [ "$$CERT_EXISTS" -eq 1 ] && [ "$$DAYS_LEFT" -gt 30 ]; then \
	  echo "Certificate for $(DOMAIN) is still valid ($$DAYS_LEFT days). Skipping issuance."; \
	  echo "To force renewal run: make ssl-renew DOMAIN=$(DOMAIN)"; \
	else \
	  echo "Issuing new certificate for $(DOMAIN)..."; \
	  docker compose -f docker-compose.production.yml run --rm certbot \
	    certonly --webroot -w /var/www/certbot \
	    -d $(DOMAIN) \
	    --email $(EMAIL) \
	    --agree-tos --no-eff-email \
	    --non-interactive; \
	  echo "Reloading Nginx..."; \
	  docker exec suberes_nginx nginx -s reload; \
	fi

# ── SSL: force renew (misal hampir expired) ───────────────────────────────────
ssl-renew:
	docker compose -f docker-compose.production.yml run --rm certbot renew \
	    --cert-name $(DOMAIN) --non-interactive
	docker exec suberes_nginx nginx -s reload

# ── Deployment ────────────────────────────────────────────────────────────────
deploy:
	IMAGE_TAG=$(IMAGE_TAG) bash deploy/deploy.sh

# ── Database backup ───────────────────────────────────────────────────────────
backup:
	bash deploy/backup.sh

# ── Database seed (run AFTER app started and GORM auto-migrated) ──────────────
# ENV=staging  → uses suberes_postgres_stag container and STAG_DATABASE env var
# ENV=production (default) → uses suberes_postgres container and PROD_DATABASE
db-seed:
	$(eval CONTAINER=$(if $(filter staging,$(ENV)),suberes_postgres_stag,suberes_postgres))
	$(eval COMPOSE_FILE=$(if $(filter staging,$(ENV)),docker-compose.staging.yml,docker-compose.production.yml))
	$(eval ENV_FILE=$(if $(filter staging,$(ENV)),.env.staging,.env.production))
	$(eval DB_USER=$(shell grep '^PROD_USERNAME\|^STAG_USERNAME' $(ENV_FILE) | head -1 | cut -d= -f2 | tr -d ' '))
	$(eval DB_NAME=$(shell grep '^PROD_DATABASE\|^STAG_DATABASE' $(ENV_FILE) | head -1 | cut -d= -f2 | tr -d ' '))
	@echo "==> Seeding $(CONTAINER) → $(DB_NAME) as $(DB_USER)"
	docker exec -i $(CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < deploy/seed.sql

.PHONY: dev stag prod build test \
        docker-build docker-staging docker-staging-down \
        docker-production docker-production-down docker-logs \
        ssl-check ssl-init ssl-renew deploy backup db-seed
