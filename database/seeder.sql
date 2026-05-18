-- Seeder DML insertions for MulticamObserver

-- 1. Seed default admin credentials (Email: admin@multicamobserver.com / Password: ObserverAdmin2026!)
INSERT INTO users (email, password_hash)
VALUES ('admin@multicamobserver.com', '$2a$10$xGLnsiMdhS5tcLijVlPjgePiF0E8CuNr/9nRMCgfPeoIr9yH93hIq')
ON CONFLICT (email) DO NOTHING;

-- 2. Seed default broadcaster camera credentials
-- Node 1: Workspace Camera (ID: cam-workspace / Password: CameraNodeSecure1!)
INSERT INTO broadcasters (node_id, name, password_hash)
VALUES ('cam-workspace', 'Workspace Camera', '$2a$10$I0L2g.jMs01vQ4DBiZomc.THjh6TZuhdAVJgQI15BvayL9UpEQ47q')
ON CONFLICT (node_id) DO NOTHING;
