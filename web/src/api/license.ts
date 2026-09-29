import api from './client'
import type { ApiResponse, UserEntitlements } from './types'

export const licenseApi = {
  // Readable by every signed-in user; the full license view stays under /admin/license.
  entitlements: () => api.get<ApiResponse<UserEntitlements>>('/system/license/entitlements'),
}
