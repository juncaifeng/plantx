package domain

// ResourceStatus represents the lifecycle status of a registry resource.
type ResourceStatus string

const (
	ResourceStatusDraft    ResourceStatus = "DRAFT"
	ResourceStatusPending  ResourceStatus = "PENDING"
	ResourceStatusOnline   ResourceStatus = "ONLINE"
	ResourceStatusOffline  ResourceStatus = "OFFLINE"
	ResourceStatusUpdating ResourceStatus = "UPDATING"
)

var menuStatusTransitions = map[ResourceStatus][]ResourceStatus{
	ResourceStatusDraft:    {ResourceStatusPending},
	ResourceStatusPending:  {ResourceStatusOnline, ResourceStatusOffline, ResourceStatusDraft},
	ResourceStatusOnline:   {ResourceStatusUpdating, ResourceStatusOffline},
	ResourceStatusUpdating: {ResourceStatusOnline, ResourceStatusOffline},
	ResourceStatusOffline:  {ResourceStatusPending, ResourceStatusDraft},
}

// CanTransitionMenu returns true if a menu is allowed to transition from one
// status to another according to the lifecycle state machine.
func CanTransitionMenu(from, to ResourceStatus) bool {
	if from == "" || from == to {
		return true
	}
	allowed, ok := menuStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
