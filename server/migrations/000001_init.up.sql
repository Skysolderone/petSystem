CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20) UNIQUE,
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    avatar_url VARCHAR(500) NOT NULL DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    wechat_open_id VARCHAR(100) UNIQUE,
    apple_id VARCHAR(100) UNIQUE,
    google_id VARCHAR(100) UNIQUE,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    plan_type VARCHAR(20) NOT NULL DEFAULT 'free',
    plan_expiry TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

CREATE TABLE IF NOT EXISTS pets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    species VARCHAR(20) NOT NULL,
    breed VARCHAR(50) NOT NULL DEFAULT '',
    gender VARCHAR(10) NOT NULL DEFAULT '',
    birth_date TIMESTAMPTZ,
    weight DOUBLE PRECISION,
    avatar_url VARCHAR(500) NOT NULL DEFAULT '',
    microchip VARCHAR(50),
    is_neutered BOOLEAN NOT NULL DEFAULT FALSE,
    allergies JSONB NOT NULL DEFAULT '[]'::jsonb,
    notes TEXT NOT NULL DEFAULT '',
    health_score INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pets_owner_id ON pets (owner_id);
CREATE INDEX IF NOT EXISTS idx_pets_deleted_at ON pets (deleted_at);
