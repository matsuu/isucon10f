#!/bin/sh

#set -e

sudo rm -f /var/log/nginx/access.log
sudo rm -f /var/log/nginx/error.log
sudo systemctl reload nginx

sudo rm -f /var/log/mysql/error.log
sudo rm -f /var/log/mysql/mysql-slow.log
sudo mysqladmin flush-logs

#for ip in s2 s3 ; do
#  ssh $ip sudo rm -f /var/log/mysql/error.log
#  ssh $ip sudo rm -f /var/log/mysql/mysql-slow.log
#  ssh $ip sudo mysqladmin flush-logs
#done

sudo systemctl restart isuumo.go
sudo journalctl -u isuumo.go -f
