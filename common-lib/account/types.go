package account

// Role defines the access level of an account.
type Role string

// AccountStatus defines the current state of an account.
type AccountStatus string

const (
	// RoleAdmin represents elevated privileges.
	RoleAdmin Role = "ADMIN"
	// RoleMember is the standard user role.
	RoleMember Role = "MEMBER"

	// StatusActive means the account is usable.
	StatusActive AccountStatus = "ACTIVE"
	// StatusInactive blocks the account until activation.
	StatusInactive AccountStatus = "INACTIVE"
)
