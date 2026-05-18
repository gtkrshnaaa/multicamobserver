-- Seeder: 0001_seed_credentials
-- Target: MulticamObserver DML Insertions

-- 1. Seed default admin credentials (Username: admin / Email: admin@multicamobserver.com / Password: ObserverAdmin2026!)
-- Hashed automatically at database level using pgcrypto blowfish crypt
INSERT INTO users (username, email, password_hash)
VALUES ('admin', 'admin@multicamobserver.com', crypt('ObserverAdmin2026!', gen_salt('bf')))
ON CONFLICT (username) DO NOTHING;

-- 2. Seed default broadcaster camera credentials
-- Username: workspace_camera / Name: Workspace Camera / Password: CameraNodeSecure1!
-- Hashed automatically at database level using pgcrypto blowfish crypt
INSERT INTO broadcasters (username, name, password_hash)
VALUES ('workspace_camera', 'Workspace Camera', crypt('CameraNodeSecure1!', gen_salt('bf')))
ON CONFLICT (username) DO NOTHING;
