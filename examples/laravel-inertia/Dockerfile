# Production image following serversideup/php packaging guidance:
# https://serversideup.net/open-source/docker-php/docs/deployment-and-production/packaging-your-app-for-deployment
#
# The FrankenPHP image already includes composer, PHP CLI, pdo_pgsql, and
# install-php-extensions. It does NOT include Node, npm, or Bun — front-end
# assets are built in a separate stage and copied in.
#
# Runtime behaviour (SSL, health path, AUTORUN, PHP ini, APP_*) is NOT baked
# here. Those are environment variables applied when the container starts
# (see envs/*/stack.yaml).

# ── Frontend (npm is not in serversideup/php) ────────────────────────────────
FROM node:22-alpine AS assets
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build

# ── Runtime ──────────────────────────────────────────────────────────────────
# Pin a version. Prefer the Debian tag; upstream notes alpine FrankenPHP
# performance issues. Tag without -alpine is Debian.
FROM serversideup/php:8.4-frankenphp

# Root only to add extensions that are not in the image. pdo_pgsql, redis,
# zip, opcache, and pcntl already ship. intl is the usual Laravel extra.
USER root
RUN install-php-extensions intl
USER www-data

WORKDIR /var/www/html

# composer is already on PATH and COMPOSER_HOME is writable by www-data.
COPY --chown=www-data:www-data composer.json composer.lock ./
RUN composer install \
      --no-dev \
      --no-interaction \
      --no-scripts \
      --prefer-dist \
      --optimize-autoloader

COPY --chown=www-data:www-data . .
RUN composer dump-autoload --optimize --no-dev \
 && mkdir -p storage/app/public bootstrap/cache

COPY --from=assets --chown=www-data:www-data /app/public/build ./public/build
