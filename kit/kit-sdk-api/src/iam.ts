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

export interface Attribute {
  id: string;
  key: string;
  valueType: string;
  description?: string;
}

export interface Condition {
  id: string;
  name: string;
  attributeKey: string;
  operator: string;
  value: string;
  description?: string;
}

export interface Policy {
  id: string;
  name: string;
  description?: string;
  permissions: string[];
  conditionIds: string[];
  effect: string;
  priority: number;
}

export interface ListAttributesRequest {}
export interface ListAttributesResponse {
  attributes: Attribute[];
}
export interface CreateAttributeRequest {
  key: string;
  valueType: string;
  description?: string;
}
export interface UpdateAttributeRequest {
  id: string;
  key: string;
  valueType: string;
  description?: string;
}
export interface DeleteAttributeRequest {
  id: string;
}

export interface ListConditionsRequest {}
export interface ListConditionsResponse {
  conditions: Condition[];
}
export interface CreateConditionRequest {
  name: string;
  attributeKey: string;
  operator: string;
  value: string;
  description?: string;
}
export interface UpdateConditionRequest {
  id: string;
  name: string;
  attributeKey: string;
  operator: string;
  value: string;
  description?: string;
}
export interface DeleteConditionRequest {
  id: string;
}

export interface ListPoliciesRequest {}
export interface ListPoliciesResponse {
  policies: Policy[];
}
export interface CreatePolicyRequest {
  name: string;
  description?: string;
  permissions: string[];
  conditionIds: string[];
  effect: string;
  priority: number;
}
export interface UpdatePolicyRequest {
  id: string;
  name: string;
  description?: string;
  permissions: string[];
  conditionIds: string[];
  effect: string;
  priority: number;
}
export interface DeletePolicyRequest {
  id: string;
}

export interface EvaluatePolicyRequest {
  permission: string;
  user_attributes?: Record<string, string>;
  resource_attributes?: Record<string, string>;
}
export interface EvaluatePolicyResponse {
  allowed: boolean;
  reason: string;
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

  listAttributes(): Promise<ListAttributesResponse> {
    return this.client.get<ListAttributesResponse>('iam/v1/attributes');
  }
  createAttribute(req: CreateAttributeRequest): Promise<Attribute> {
    return this.client.post<Attribute>('iam/v1/attributes', req);
  }
  updateAttribute(req: UpdateAttributeRequest): Promise<Attribute> {
    return this.client.put<Attribute>(`iam/v1/attributes/${encodeURIComponent(req.id)}`, req);
  }
  deleteAttribute(req: DeleteAttributeRequest): Promise<void> {
    return this.client.delete<void>(`iam/v1/attributes/${encodeURIComponent(req.id)}`);
  }

  listConditions(): Promise<ListConditionsResponse> {
    return this.client.get<ListConditionsResponse>('iam/v1/conditions');
  }
  createCondition(req: CreateConditionRequest): Promise<Condition> {
    return this.client.post<Condition>('iam/v1/conditions', req);
  }
  updateCondition(req: UpdateConditionRequest): Promise<Condition> {
    return this.client.put<Condition>(`iam/v1/conditions/${encodeURIComponent(req.id)}`, req);
  }
  deleteCondition(req: DeleteConditionRequest): Promise<void> {
    return this.client.delete<void>(`iam/v1/conditions/${encodeURIComponent(req.id)}`);
  }

  listPolicies(): Promise<ListPoliciesResponse> {
    return this.client.get<ListPoliciesResponse>('iam/v1/policies');
  }
  createPolicy(req: CreatePolicyRequest): Promise<Policy> {
    return this.client.post<Policy>('iam/v1/policies', req);
  }
  updatePolicy(req: UpdatePolicyRequest): Promise<Policy> {
    return this.client.put<Policy>(`iam/v1/policies/${encodeURIComponent(req.id)}`, req);
  }
  deletePolicy(req: DeletePolicyRequest): Promise<void> {
    return this.client.delete<void>(`iam/v1/policies/${encodeURIComponent(req.id)}`);
  }
  evaluatePolicy(req: EvaluatePolicyRequest): Promise<EvaluatePolicyResponse> {
    return this.client.post<EvaluatePolicyResponse>('iam/v1/policies/evaluate', req);
  }
}
