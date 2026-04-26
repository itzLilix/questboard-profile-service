package usecase

type ViewerContext struct {
	UserID  string
	IsAdmin bool
}

func (v *ViewerContext) IsAuthenticated() bool {
	return v != nil && v.UserID != ""
}