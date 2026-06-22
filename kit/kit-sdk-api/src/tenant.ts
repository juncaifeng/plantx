import { KitApiClient } from './index.js';

export interface Tenant {
  id: string;
  name: string;
  status: string;
  created_at: number;
}

export interface ListTenantsRequest {}

export interface ListTenantsResponse {
  tenants: Tenant[];
}

export interface CreateTenantRequest {
  name: string;
}

export class TenantServiceClient {
  constructor(private readonly client: KitApiClient) {}

  listTenants(_request?: ListTenantsRequest): Promise<ListTenantsResponse> {
    return this.client.get<ListTenantsResponse>('tenant/v1/tenants');
  }

  createTenant(request: CreateTenantRequest): Promise<Tenant> {
    return this.client.post<Tenant>('tenant/v1/tenants', request);
  }
}
