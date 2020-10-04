#!/bin/sh

. /home/isucon/env

mysql -h ${MYSQL_HOSTNAME} -P ${MYSQL_PORT} -u ${MYSQL_USER} -p${MYSQL_PASS} ${MYSQL_DATABASE}
