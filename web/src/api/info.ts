import api from './client'
import type { ApiResponse, AppInfo, ComponentInfo } from './types'

export const infoApi = {
  get: () => api.get<ApiResponse<AppInfo>>('/info'),
  // Signed-in users only; the About page's Components card.
  components: () => api.get<ApiResponse<ComponentInfo[]>>('/system/components'),
}
