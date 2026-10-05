# Sample multi-stage image for Laravel + React (Inertia / Vite) on
# serversideup/php FrankenPHP — the variation recommended for new Laravel apps.
# Docs: https://serversideup.net/open-source/docker-php/docs/image-variations/frankenphp
#
# Stages:
#   1. composer  — PHP dependencies
#   2. frontend  — npm ci + vite build (Inertia React)
#   3. runtime   — FrankenPHP (Caddy + PHP in one process) on :8080

# ── 1. Composer ──────────────────────────────────────────────────────────────
FROM composer:2 AS vendor
WORKDIR /app
COPY composer.json composer.lock ./
RUN composer install \
      --no-dev \
      --no-interaction \
      --no-scripts \
      --prefer-dist \
      --optimize-autoloader
COPY . .
RUN composer dump-autoload --optimize --no-dev

# ── 2. Frontend (Vite / Inertia React) ───────────────────────────────────────
FROM node:22-alpine AS frontend
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
COPY --from=vendor /app/vendor ./vendor
RUN npm run build

# ── 3. Runtime (FrankenPHP) ──────────────────────────────────────────────────
# Unprivileged www-data; HTTP on 8080 (Miabi / Goma terminate TLS upstream).
# Prefer debian over alpine if you hit FrankenPHP alpine perf issues.
FROM serversideup/php:8.4-frankenphp
WORKDIR /var/www/html

USER root
COPY --chown=www-data:www-data . .
COPY --from=vendor --chown=www-data:www-data /app/vendor ./vendor
COPY --from=frontend --chown=www-data:www-data /app/public/build ./public/build

# Uploads volume mounts at storage/app/public; AUTORUN also runs storage:link.
RUN mkdir -p storage/app/public bootstrap/cache \
 && chown -R www-data:www-data storage bootstrap/cache

USER www-data

# Classic FrankenPHP mode (safe default). For Octane worker mode, override the
# Application command to:
#   ["php", "artisan", "octane:start", "--server=frankenphp", "--port=8080"]
# and switch healthcheck to the image's healthcheck-octane (see README).
ENV APP_ENV=production \
    LOG_CHANNEL=stderr \
    SSL_MODE=off \
    CADDY_SERVER_ROOT=/var/www/html/public \
    HEALTHCHECK_PATH=/up
EXPOSE 8080
