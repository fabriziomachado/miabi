import { ref } from 'vue'
import { infoApi } from '@/api/info'
import type { ComponentInfo } from '@/api/types'

// One fetch per page load, shared by every beta banner and the About page's table.
const components = ref<ComponentInfo[]>([])
let pending: Promise<void> | null = null

export function useComponents() {
  if (!pending) {
    pending = infoApi
      .components()
      .then((res) => { components.value = res.data.data })
      .catch(() => { pending = null }) // retried by the next caller
  }
  const byId = (id: string) => components.value.find((c) => c.id === id)
  return { components, byId }
}
