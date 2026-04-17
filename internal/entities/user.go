package entities

import (
	"time"

	"github.com/itzLilix/questboard-shared/dtos"
)

type User struct {
	ID           	string   	`db:"id"`
	Username     	string   	`db:"username"`
	PasswordHash 	string   	`db:"password_hash"`
	Email        	string   	`db:"email"`
	CreatedAt    	time.Time  	`db:"created_at"`
	LastLogin    	time.Time  	`db:"last_login"`
	AvatarURL    	*string   	`db:"avatar_url"`
	BannerURL    	*string   	`db:"banner_url"`
	Role         	dtos.Role  	`db:"role"`
	DisplayName  	string		`db:"display_name"`
	IsEmailVerified bool 		`db:"is_email_verified"`
	SessionsPlayed 	int			`db:"sessions_played"`
	SessionsHosted 	int			`db:"sessions_hosted"`
	Rating 			float64		`db:"rating"`
	ReviewsCount    int			`db:"reviews_count"`
	Bio				*string		`db:"bio"`
	Links			[]dtos.Link	`db:"links"`
}