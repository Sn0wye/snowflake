-- Create 'helium' database if it does not exist
SELECT 'CREATE DATABASE helium'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'helium'
)\gexec

-- Create 'carbon' database if it does not exist
SELECT 'CREATE DATABASE carbon'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'carbon'
)\gexec

-- Create 'loan' database if it does not exist
SELECT 'CREATE DATABASE loan'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'loan'
)\gexec

-- Create 'gold' database if it does not exist
SELECT 'CREATE DATABASE gold'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'gold'
)\gexec