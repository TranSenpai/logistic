#!/bin/bash
set -e

echo "Waiting for master to be ready..."
# Wait until mysql master is ready
while ! mysql -h"$MASTER_HOST" -u"root" -p"$MYSQL_ROOT_PASSWORD" -e "status" > /dev/null 2>&1; do
    echo "Master not ready yet, waiting 5 seconds..."
    sleep 5
done

echo "Master is ready. Setting up replication..."

mysql -u"root" -p"$MYSQL_ROOT_PASSWORD" <<-EOSQL
  CHANGE REPLICATION SOURCE TO
    SOURCE_HOST='$MASTER_HOST',
    SOURCE_USER='$REPLICATION_USER',
    SOURCE_PASSWORD='$REPLICATION_PASSWORD',
    SOURCE_PORT=3306,
    GET_SOURCE_PUBLIC_KEY=1;
  START REPLICA;
EOSQL

echo "Replication started."
