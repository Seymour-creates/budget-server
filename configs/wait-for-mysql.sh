#!/bin/sh
# wait-for-mysql.sh

set -e

host="$1"
shift
cmd=""
for arg in "$@"; do
    cmd="$cmd \"$arg\""
done

until mysql -h "$host" -u"$DB_USER" -p"$DB_PASS" -e 'SELECT 1'; do
  >&2 echo "MySQL is unavailable - sleeping"
  sleep 1
done

>&2 echo "MySQL is up - executing command"
eval exec "$cmd"