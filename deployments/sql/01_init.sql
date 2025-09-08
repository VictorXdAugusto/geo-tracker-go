CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location GEOMETRY(POINT, 4326) NOT NULL,
    sector_x INTEGER NOT NULL,
    sector_y INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_positions_user_id ON positions (user_id);
CREATE INDEX IF NOT EXISTS idx_positions_location ON positions USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_positions_sector ON positions (sector_x, sector_y);
CREATE INDEX IF NOT EXISTS idx_positions_created_at ON positions (created_at DESC);

CREATE TABLE IF NOT EXISTS current_positions (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    location GEOMETRY(POINT, 4326) NOT NULL,
    sector_x INTEGER NOT NULL,
    sector_y INTEGER NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_current_positions_location ON current_positions USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_current_positions_sector ON current_positions (sector_x, sector_y);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_current_positions_updated_at BEFORE UPDATE ON current_positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
