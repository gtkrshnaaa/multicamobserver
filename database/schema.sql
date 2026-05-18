-- Schema definition for MulticamObserver

-- 1. Create users table for Administrator (Viewer Node)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create broadcasters table for physical camera nodes
CREATE TABLE IF NOT EXISTS broadcasters (
    id SERIAL PRIMARY KEY,
    node_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Seed default admin credentials (Email: admin@multicamobserver.com / Password: ObserverAdmin2026!)
INSERT INTO users (email, password_hash)
VALUES ('admin@multicamobserver.com', '$2a$10$xGLnsiMdhS5tcLijVlPjgePiF0E8CuNr/9nRMCgfPeoIr9yH93hIq')
ON CONFLICT (email) DO NOTHING;

-- 4. Seed default broadcaster camera credentials
-- Node 1: Workspace Camera (ID: cam-workspace / Password: CameraNodeSecure1!)
INSERT INTO broadcasters (node_id, name, password_hash)
VALUES ('cam-workspace', 'Workspace Camera', '$2a$10$I0L2g.jMs01vQ4DBiZomc.THjh6TZuhdAVJgQI15BvayL9UpEQ47q')
ON CONFLICT (node_id) DO NOTHING;

-- Node 2: Front Door Camera (ID: cam-frontdoor / Password: CameraNodeSecure2!)
INSERT INTO broadcasters (node_id, name, password_hash)
VALUES ('cam-frontdoor', 'Front Door Camera', '$2a$10$XEJpOAh6q3r2ySKvaY9HIei5BLoL74fG/XJLCJMrtzTg6hu2Ptjg6')
ON CONFLICT (node_id) DO NOTHING;
