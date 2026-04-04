#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
    CREATE DATABASE "user";
    CREATE DATABASE merchant;
    CREATE DATABASE catalog;
    CREATE DATABASE api;
EOSQL