package entities

import (
	"time"

	"github.com/itzLilix/questboard-shared/dtos"
)

type User struct {
	ID           	string   	
	Username     	string   	
	PasswordHash 	string   	
	Email        	string   	
	CreatedAt    	time.Time  	
	LastLogin    	time.Time  	
	AvatarURL    	*string   	
	BannerURL    	*string   	
	Role         	dtos.Role   
	DisplayName  	string		
	IsEmailVerified bool 		
	SessionsPlayed 	int			
	SessionsHosted 	int			
	Rating 			float64		
	ReviewsCount    int			
	Bio				*string		
	Links			[]dtos.Link	
}