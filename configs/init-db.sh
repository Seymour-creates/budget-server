#!/bin/sh
#initialize db to use passwords
mysql -u root -p"$MYSQL_ROOT_PASSWORD" <<-EOSQL
    ALTER USER 'root'@'%' IDENTIFIED WITH 'mysql_native_password' BY '$MYSQL_ROOT_PASSWORD';
#    USE budget;
#    SOURCE /docker-entrypoint-initdb.d/export.sql;
EOSQL