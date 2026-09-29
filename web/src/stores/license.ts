import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { adminApi } from '@/api/admin'
import { licenseApi } from '@/api/license'
import { useAuthStore } from '@/stores/auth'
import type { LicenseView, LicenseEdition, LicenseState, UserEntitlements } from '@/api/types'

// Caches commercial entitlements so the banner and per-feature gates
// (useEntitlement) share one source of truth. Admins load the full view; every
// other user loads the gating subset (edition, state, flags). In Community
// installs the edition is "community" and every flag is false.
export const useLicenseStore = defineStore('license', () => {
  const view = ref<LicenseView | null>(null)
  const gates = ref<UserEntitlements | null>(null)
  const loaded = ref(false)

  const current = computed<UserEntitlements | null>(() => view.value ?? gates.value)
  const edition = computed<LicenseEdition>(() => current.value?.edition ?? 'community')
  const state = computed<LicenseState>(() => current.value?.state ?? 'none')
  const warnings = computed<string[]>(() => view.value?.warnings ?? [])
  const isCommunity = computed(() => edition.value === 'community')

  // has reports whether an entitlement flag is usable at runtime (false in CE).
  function has(flag: string): boolean {
    return !!current.value?.flags?.[flag]
  }

  // mutable reports whether a flag's config may be changed: entitled AND the
  // license is not degraded past its grace period (mirrors the backend gate).
  function mutable(flag: string): boolean {
    return has(flag) && (state.value === 'valid' || state.value === 'grace')
  }

  // inflight dedupes concurrent load() calls (many admin pages call it on mount)
  // so they share one request rather than racing.
  let inflight: Promise<void> | null = null

  // loadedAsAdmin records which view is cached, so a sign-in as another role in the same tab reloads.
  let loadedAsAdmin: boolean | null = null

  async function load(force = false): Promise<void> {
    const isAdmin = useAuthStore().isAdmin
    if (loaded.value && loadedAsAdmin === isAdmin && !force) return
    if (inflight && !force) return inflight
    // /admin/license is admin-only; a non-admin asking for it got a 403 and saw every EE feature locked.
    const request = isAdmin
      ? adminApi.getLicense().then((res) => { view.value = res.data.data; gates.value = null })
      : licenseApi.entitlements().then((res) => { gates.value = res.data.data; view.value = null })
    inflight = request
      .then(() => {
        loadedAsAdmin = isAdmin
        loaded.value = true // only mark loaded on SUCCESS
      })
      .catch(() => {
        // A transient failure (network/500/token-not-ready) must NOT permanently
        // disable entitlements: keep any prior view and leave `loaded` false so the
        // next gate access retries, instead of sticking every flag at "off".
      })
      .finally(() => {
        inflight = null
      })
    return inflight
  }

  // set updates the cache from a fresh view (e.g. right after install/remove) so
  // the banner reacts immediately without a refetch.
  function set(v: LicenseView | null): void {
    view.value = v
    loaded.value = true
    loadedAsAdmin = true
  }

  return { view, gates, loaded, edition, state, warnings, isCommunity, has, mutable, load, set }
})
