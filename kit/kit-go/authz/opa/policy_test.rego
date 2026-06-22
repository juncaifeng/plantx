package plantx.authz

import rego.v1

default allow := false

allow if {
	input.user.roles[_] == "admin"
}

allow if {
	perm := input.user.permissions[_]
	perm == concat(":", [input.action.service, input.action.resource, input.action.operation])
}
