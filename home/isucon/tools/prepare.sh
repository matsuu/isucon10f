#!/bin/sh

#set -e

#sudo rm -f /var/log/nginx/access.log
#sudo rm -f /var/log/nginx/error.log
sudo rm -f /var/log/envoy/access.log
sudo systemctl restart envoy
sudo rm -f /var/log/varnish/varnishncsa.log
sudo systemctl reload varnishncsa

#sudo rm -f /var/log/mysql/error.log
#sudo rm -f /var/log/mysql/mysql-slow.log
#sudo mysqladmin flush-logs

for ip in isu2.t.isucon.dev ; do
  ssh $ip sudo rm -f /var/log/mysql/error.log
  ssh $ip sudo rm -f /var/log/mysql/mysql-slow.log
  ssh $ip sudo mysqladmin flush-logs
done

#sudo systemctl restart xsuportal-web-golang.service
#sudo systemctl restart xsuportal-api-golang.service

for ip in isu3.t.isucon.dev ; do
  ssh $ip sudo systemctl restart xsuportal-web-golang.service
  ssh $ip sudo systemctl restart xsuportal-api-golang.service
done

#sudo journalctl -u xsuportal-web-golang.service -f
sudo systemctl restart varnish
