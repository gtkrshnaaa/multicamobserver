-- Seeder: 0001_seed_credentials
-- Target: MulticamObserver DML Insertions

-- 1. Seed default admin credentials (Email: admin@multicamobserver.com / Password: ObserverAdmin2026!)
-- Hashed automatically at database level using pgcrypto blowfish crypt
INSERT INTO users (email, password_hash)
VALUES ('admin@multicamobserver.com', crypt('ObserverAdmin2026!', gen_salt('bf')))
ON CONFLICT (email) DO NOTHING;

-- 2. Seed default broadcaster camera credentials
-- Password: CameraNodeSecure1! (Node ID is auto-generated via UUID)
-- Hashed automatically at database level using pgcrypto blowfish crypt
INSERT INTO broadcasters (name, password_hash)
VALUES ('Workspace Camera', crypt('CameraNodeSecure1!', gen_salt('bf')));
