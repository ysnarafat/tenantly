-- Initialize database with proper settings
CREATE DATABASE tenantly;

-- Connect to tenantly database
\c tenantly;

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'Asia/Dhaka';