ALTER TABLE users
ADD COLUMN preferred_type   varchar(10) NULL
    CHECK (preferred_type IS NULL OR preferred_type IN ('oneshot','campaign')),
ADD COLUMN preferred_format varchar(10) NULL
    CHECK (preferred_format IS NULL OR preferred_format IN ('online','offline')),
ADD COLUMN city varchar(100) NULL,
ADD COLUMN is_visible_in_catalog boolean NOT NULL DEFAULT true;

CREATE INDEX users_city_lower_idx ON users (lower(city)) WHERE city IS NOT NULL;