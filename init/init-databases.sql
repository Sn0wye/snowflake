-- Create 'authentication' database if it does not exist
SELECT 'CREATE DATABASE authentication'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'authentication'
)\gexec

-- Create 'scorer' database if it does not exist
SELECT 'CREATE DATABASE scorer'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'scorer'
)\gexec

-- Create 'loan' database if it does not exist
SELECT 'CREATE DATABASE loan'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'loan'
)\gexec

-- Create 'transaction' database if it does not exist
SELECT 'CREATE DATABASE core'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'transaction'
)\gexec