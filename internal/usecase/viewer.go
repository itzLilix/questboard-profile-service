package usecase

import "github.com/itzLilix/questboard-shared/dtos"

type Viewer struct {
	UserID string
	Role   dtos.Role
}

func (v *Viewer) IsAuthenticated() bool { return v != nil && v.UserID != "" }
func (v *Viewer) IsAdmin() bool         { return v != nil && v.Role == dtos.AdminRole }
func (v *Viewer) Is(userID string) bool { return v != nil && v.UserID == userID }
func (v *Viewer) CanActAs(ownerID string) bool { // owner or admin
    return v.IsAuthenticated() && (v.UserID == ownerID || v.IsAdmin())
}