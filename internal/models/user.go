package models

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           	string   	`json:"id"`
	Username     	string   	`json:"username"`
	PasswordHash 	string   	`json:"-"`
	Email        	string   	`json:"email"`
	CreatedAt    	time.Time  	`json:"createdAt"`
	LastLogin    	*time.Time  `json:"lastLogin,omitempty"`
	AvatarURL    	*string   	`json:"avatarUrl,omitempty"`
	BannerURL    	*string   	`json:"bannerUrl,omitempty"`
	Role         	string   	`json:"role"`
	DisplayName  	string		`json:"displayName"`
	IsEmailVerified bool 		`json:"isEmailVerified"`
	SessionsPlayed 	int			`json:"sessionsPlayed"`
	SessionsHosted 	int			`json:"sessionsHosted"`
	Rating 			float64		`json:"rating"`
	ReviewsCount    int			`json:"reviewsCount"`
}

type RefreshToken struct {
	ID string `json:"id"`
	UserID string `json:"userId"`
	TokenPrefix string `json:"tokenPrefix"`
	TokenHash string `json:"tokenHash"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

func ScanUser(row pgx.Row, user *User) error{
	return row.Scan(
		&user.ID, 
		&user.Username, 
		&user.PasswordHash, 
		&user.Email, 
		&user.CreatedAt, 
		&user.LastLogin, 
		&user.AvatarURL, 
		&user.BannerURL, 
		&user.Role, 
		&user.DisplayName, 
		&user.IsEmailVerified,
		&user.SessionsPlayed,
		&user.SessionsHosted,
		&user.Rating,
		&user.ReviewsCount)
}