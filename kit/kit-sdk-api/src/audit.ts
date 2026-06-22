import type { KitApiClient } from './index.js';

export interface AuditLog {
  id: string;
  tenant_id: string;
  user_id: string;
  action: string;
  resource: string;
  timestamp: number | string;
  detail: string;
}

export interface ListAuditLogsRequest {
  tenant_id?: string;
  start_time?: number | string;
  end_time?: number | string;
  limit?: number;
}

export interface ListAuditLogsResponse {
  logs: AuditLog[];
}

export class AuditServiceClient {
  constructor(private readonly client: KitApiClient) {}

  async listAuditLogs(req: ListAuditLogsRequest = {}): Promise<ListAuditLogsResponse> {
    const params = new URLSearchParams();
    if (req.tenant_id) {
      params.set('tenant_id', String(req.tenant_id));
    }
    if (req.start_time !== undefined) {
      params.set('start_time', String(req.start_time));
    }
    if (req.end_time !== undefined) {
      params.set('end_time', String(req.end_time));
    }
    if (req.limit !== undefined) {
      params.set('limit', String(req.limit));
    }
    const query = params.toString();
    const path = query ? `/audit/v1/logs?${query}` : '/audit/v1/logs';
    return this.client.get<ListAuditLogsResponse>(path);
  }
}
