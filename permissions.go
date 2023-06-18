package revoltgo

// PermissionAD describes the default allowed and denied permissions for users
type PermissionAD struct {
	Allow uint `json:"a"`
	Deny  uint `json:"d"`
}
