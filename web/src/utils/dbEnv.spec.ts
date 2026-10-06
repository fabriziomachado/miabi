import { describe, expect, it } from 'vitest'
import { envFields, prefixedName, sanitizePrefix } from './dbEnv'

describe('dbEnv', () => {
  it('normalizes prefixes like the server', () => {
    expect(sanitizePrefix(' my-cache ')).toBe('MY_CACHE')
    expect(sanitizePrefix('_a.b_')).toBe('AB')
  })

  it('prefixes default names per engine', () => {
    expect(prefixedName('postgres', '', 'url')).toBe('DB_URL')
    expect(prefixedName('postgres', 'ANALYTICS', 'database_url')).toBe('ANALYTICS_DATABASE_URL')
    expect(prefixedName('redis', 'CACHE', 'url')).toBe('CACHE_REDIS_URL')
  })

  it('omits fields an engine has no value for', () => {
    expect(envFields('redis')).toEqual(['url', 'host', 'port', 'password'])
  })
})
