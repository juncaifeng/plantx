package plantx.authz

import rego.v1

default allow := false

# Admin can do anything
allow if "admin" in input.user.roles

# Permission-based authorization within allowed tenant scope
allow if {
	required := sprintf("%s:%s", [input.action.resource, input.action.operation])
	input.user.permissions[_] == required
	tenant_ok
}

# Tenant scope is OK when no resource tenant is specified or when it matches the user tenant
tenant_ok if not input.resource.tenant_id

tenant_ok if input.resource.tenant_id == ""

tenant_ok if input.resource.tenant_id == input.user.tenant_id
