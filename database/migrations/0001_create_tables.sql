-- Migration: 0001_create_tables
-- Target: MulticamObserver DDL Structures

-- Enable pgcrypto extension for database-level bcrypt hashing
CREATE EXTENSION IF NOT EXISTS pgcrypto;

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
    node_id VARCHAR(100) UNIQUE NOT NULL DEFAULT gen_random_uuid()::text,
    name VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
