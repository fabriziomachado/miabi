import api from './client'
import type { ApiResponse } from './types'

// Runner is a dedicated build/pipeline machine (workspace-owned or, when
// workspace_id is null, a platform-shared runner in the admin pool).
export interface Runner {
  id: number
  uid: string
  name: string
  display_name: string
  workspace_id: number | null
  scope: 'workspace' | 'shared'
  labels: string[] | null
  concurrency: number
  os?: string
  arch?: string
  version?: string
  // features the runner reports supporting, such as "multi-platform"
  features?: string[]
  remote_ip?: string
  status: 'online' | 'offline' | 'draining'
  cordoned: boolean
  enabled: boolean
  ephemeral: boolean
  connected: boolean
  last_seen_at?: string | null
  created_at: string
  updated_at: string
}

export interface RunnerInput {
  name: string
  display_name?: string
  labels?: string[]
  concurrency?: number
}

// Create returns the runner plus its one-time registration token (shown once)
// and the server-configured runner image for the enrollment command.
export interface RunnerCreated {
  runner: Runner
  token: string
  image: string
}

// RunnerAdapter binds either the workspace or admin API so the RunnersPanel is
// reused for both the workspace-owned runners and the platform-shared pool.
export interface RunnerAdapter {
  list: () => Promise<Runner[]>
  get: (id: number) => Promise<Runner>
  create: (input: RunnerInput) => Promise<RunnerCreated>
  cordon: (id: number, cordoned: boolean) => Promise<Runner>
  regenerateToken: (id: number) => Promise<string>
  remove: (id: number) => Promise<void>
}

const wsBase = (ws: number) => `/workspaces/${ws}/runners`

export const runnerApi = {
  list: (ws: number) => api.get<ApiResponse<Runner[]>>(wsBase(ws)),
  // Read-only view of the platform-shared pool available alongside this workspace's runners.
  listShared: (ws: number) => api.get<ApiResponse<Runner[]>>(`${wsBase(ws)}/shared`),
  get: (ws: number, id: number) => api.get<ApiResponse<Runner>>(`${wsBase(ws)}/${id}`),
  create: (ws: number, input: RunnerInput) => api.post<ApiResponse<RunnerCreated>>(wsBase(ws), input),
  update: (ws: number, id: number, input: RunnerInput) => api.put<ApiResponse<Runner>>(`${wsBase(ws)}/${id}`, input),
  cordon: (ws: number, id: number, cordoned: boolean) =>
    api.post<ApiResponse<Runner>>(`${wsBase(ws)}/${id}/cordon`, { cordoned }),
  regenerateToken: (ws: number, id: number) =>
    api.post<ApiResponse<{ token: string }>>(`${wsBase(ws)}/${id}/token`, {}),
  remove: (ws: number, id: number) => api.delete<ApiResponse<{ message: string }>>(`${wsBase(ws)}/${id}`),
}

const adminBase = '/admin/runners'

export const adminRunnerApi = {
  list: () => api.get<ApiResponse<Runner[]>>(adminBase),
  get: (id: number) => api.get<ApiResponse<Runner>>(`${adminBase}/${id}`),
  create: (input: RunnerInput) => api.post<ApiResponse<RunnerCreated>>(adminBase, input),
  update: (id: number, input: RunnerInput) => api.put<ApiResponse<Runner>>(`${adminBase}/${id}`, input),
  cordon: (id: number, cordoned: boolean) => api.post<ApiResponse<Runner>>(`${adminBase}/${id}/cordon`, { cordoned }),
  regenerateToken: (id: number) => api.post<ApiResponse<{ token: string }>>(`${adminBase}/${id}/token`, {}),
  remove: (id: number) => api.delete<ApiResponse<{ message: string }>>(`${adminBase}/${id}`),
}
