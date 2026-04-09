package models

import (
	"time"
)

type Role string

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
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
	Role         	Role   		`json:"role"`
	DisplayName  	string		`json:"displayName"`
	IsEmailVerified bool 		`json:"isEmailVerified"`
	SessionsPlayed 	int			`json:"sessionsPlayed"`
	SessionsHosted 	int			`json:"sessionsHosted"`
	Rating 			float64		`json:"rating"`
	ReviewsCount    int			`json:"reviewsCount"`
}