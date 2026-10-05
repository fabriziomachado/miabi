# Sample multi-stage image for Laravel + React (Inertia / Vite).
# Illustrative only — adjust PHP version, document root, and the web server
# to match your project. The pipeline's `uses: build` step points here.
#
# Stages:
#   1. composer  — PHP dependencies
#   2. frontend  — npm ci + vite build (Inertia React)
#   3. runtime   — PHP-FPM + Nginx serving public/ with compiled assets

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

# ── 3. Runtime ───────────────────────────────────────────────────────────────
# serversideup/php ships Nginx + PHP-FPM and listens on 8080 by default —
# matches ports.container / healthcheck in the GitOps stacks.
FROM serversideup/php:8.4-fpm-nginx
WORKDIR /var/www/html

USER root
COPY --chown=www-data:www-data . .
COPY --from=vendor --chown=www-data:www-data /app/vendor ./vendor
COPY --from=frontend --chown=www-data:www-data /app/public/build ./public/build

# Uploads volume mounts at storage/app/public; ensure the symlink target exists.
RUN mkdir -p storage/app/public bootstrap/cache \
 && php artisan storage:link || true \
 && chown -R www-data:www-data storage bootstrap/cache

USER www-data
ENV APP_ENV=production \
    LOG_CHANNEL=stderr
EXPOSE 8080
