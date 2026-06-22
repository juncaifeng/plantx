import type { KitApiClient } from './index.js';

export interface User {
  id: string;
  username: string;
  tenant_id: string;
  role_ids?: string[];
}

export interface Role {
  id: string;
  name: string;
  permissions?: string[];
  description?: string;
}

export interface Permission {
  id: string;
  name: string;
  resource: string;
  operation: string;
  description?: string;
}

export interface ListUsersRequest {}

export interface ListUsersResponse {
  users: User[];
}

export interface CreateUserRequest {
  username: string;
  tenant_id: string;
  role_ids?: string[];
}

export interface ListRolesRequest {}

export interface ListRolesResponse {
  roles: Role[];
}

export interface GetRoleRequest {
  id: string;
}

export interface CreateRoleRequest {
  name: string;
  permissions?: string[];
  description?: string;
}

export interface UpdateRoleRequest {
  id: string;
  name?: string;
  permissions?: string[];
  description?: string;
}

export interface DeleteRoleRequest {
  id: string;
}

export interface ListPermissionsRequest {}

export interface ListPermissionsResponse {
  permissions: Permission[];
}

export interface CreatePermissionRequest {
  name: string;
  resource: string;
  operation: string;
  description?: string;
}

export interface DeletePermissionRequest {
  id: string;
}

export class IAMServiceClient {
  constructor(private client: KitApiClient) {}

  listUsers(): Promise<ListUsersResponse> {
    return this.client.get<ListUsersResponse>('iam/v1/users');
  }

  createUser(req: CreateUserRequest): Promise<User> {
    return this.client.post<User>('iam/v1/users', req);
  }

  listRoles(): Promise<ListRolesResponse> {
    return this.client.get<ListRolesResponse>('iam/v1/roles');
  }

  getRole(req: GetRoleRequest): Promise<Role> {
    return this.client.get<Role>(`iam/v1/roles/${encodeURIComponent(req.id)}`);
  }

  createRole(req: CreateRoleRequest): Promise<Role> {
    return this.client.post<Role>('iam/v1/roles', req);
  }

  updateRole(req: UpdateRoleRequest): Promise<Role> {
    return this.client.put<Role>(`iam/v1/roles/${encodeURIComponent(req.id)}`, req);
  }

  deleteRole(req: DeleteRoleRequest): Promise<void> {
    return this.client.delete<void>(`iam/v1/roles/${encodeURIComponent(req.id)}`);
  }

  listPermissions(): Promise<ListPermissionsResponse> {
    return this.client.get<ListPermissionsResponse>('iam/v1/permissions');
  }

  createPermission(req: CreatePermissionRequest): Promise<Permission> {
    return this.client.post<Permission>('iam/v1/permissions', req);
  }

  deletePermission(req: DeletePermissionRequest): Promise<void> {
    return this.client.delete<void>(`iam/v1/permissions/${encodeURIComponent(req.id)}`);
  }
}
