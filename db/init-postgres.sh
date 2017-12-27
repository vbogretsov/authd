#!/bin/bash

psql -v -U $POSTGRES_USER -c "CREATE USER $DBUSER;"
psql -v -U $POSTGRES_USER -c "CREATE DATABASE $DBNAME;"
psql -v -U $POSTGRES_USER -c "GRANT ALL PRIVILEGES ON DATABASE $DBNAME TO $DBUSER;"

sed -ri "s/#log_statement = 'none'/log_statement = 'all'/g" /var/lib/postgresql/data/postgresql.conf