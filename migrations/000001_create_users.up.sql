CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR (32) UNIQUE NOT NULL CHECK (username ~ '^[a-zA-Z0-9_\.\-]+$'),
    password_hash VARCHAR (60) NOT NULL,
    email VARCHAR (355) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP,
	profile_picture TEXT,
	banner_picture TEXT,
	role VARCHAR (20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'))
)