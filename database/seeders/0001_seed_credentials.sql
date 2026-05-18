-- Seeder: 0001_seed_credentials
-- Target: MulticamObserver DML Insertions

-- 1. Seed default admin credentials (Username: admin / Password: password)
-- Hashed automatically at database level using pgcrypto blowfish crypt
INSERT INTO users (username, password_hash)
VALUES ('admin', crypt('password', gen_salt('bf')))
ON CONFLICT (username) DO NOTHING;

-- 2. Seed default broadcaster camera credentials
-- Username: camera1 / Name: Camera 1 / Password: password
-- Username: camera2 / Name: Camera 2 / Password: password
-- Hashed automatically at database level using pgcrypto blowfish crypt
INSERT INTO broadcasters (username, name, password_hash)
VALUES 
  ('camera1', 'Camera 1', crypt('password', gen_salt('bf'))),
  ('camera2', 'Camera 2', crypt('password', gen_salt('bf')))
ON CONFLICT (username) DO NOTHING;
