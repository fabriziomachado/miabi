import type { DBEngine } from '@/api/types'

// Mirrors internal/services/database/envvars.go; the server re-validates.
export const DB_ENV_FIELDS = ['url', 'database_url', 'host', 'port', 'name', 'user', 'password'] as const
export type DBEnvField = (typeof DB_ENV_FIELDS)[number]

export function defaultEnvNames(engine: DBEngine): Partial<Record<DBEnvField, string>> {
  if (engine === 'redis') {
    return { url: 'REDIS_URL', host: 'REDIS_HOST', port: 'REDIS_PORT', password: 'REDIS_PASSWORD' }
  }
  return {
    url: 'DB_URL',
    database_url: 'DATABASE_URL',
    host: 'DB_HOST',
    port: 'DB_PORT',
    name: 'DB_NAME',
    user: 'DB_USER',
    password: 'DB_PASSWORD',
  }
}

export function envFields(engine: DBEngine): DBEnvField[] {
  const defaults = defaultEnvNames(engine)
  return DB_ENV_FIELDS.filter((f) => defaults[f])
}

// Engines linked as a whole instance rather than through a logical database.
export function linksWholeInstance(engine: DBEngine): boolean {
  return engine === 'redis'
}

export function prefixedName(engine: DBEngine, prefix: string, field: DBEnvField): string {
  const def = defaultEnvNames(engine)[field] ?? ''
  return prefix ? `${prefix}_${def}` : def
}

// Normalizes a prefix the way the server does (upper-snake, trimmed underscores).
export function sanitizePrefix(s: string): string {
  return s
    .trim()
    .toUpperCase()
    .replace(/[-\s]/g, '_')
    .replace(/[^A-Z0-9_]/g, '')
    .replace(/^_+|_+$/g, '')
}
