#!/bin/sh

set -ex

(
  cd ${HOME}/webapp/golang
  make
)

#sudo rm -f /var/log/nginx/access.log
#sudo rm -f /var/log/nginx/error.log
sudo rm -f /var/log/envoy/access.log
sudo systemctl restart envoy
sudo rm -f /var/log/varnish/varnishncsa.log
#sudo systemctl reload varnishncsa

#sudo rm -f /var/log/mysql/error.log
#sudo rm -f /var/log/mysql/mysql-slow.log
#sudo mysqladmin flush-logs

for ip in isu3.t.isucon.dev ; do
  ssh $ip sudo rm -f /var/log/mysql/error.log
  ssh $ip sudo rm -f /var/log/mysql/mysql-slow.log
  ssh $ip sudo mysqladmin flush-logs
done

#sudo systemctl restart xsuportal-web-golang.service
sudo systemctl restart xsuportal-api-golang.service

for ip in isu2.t.isucon.dev
do
	ssh $ip sudo systemctl stop xsuportal-web-golang
	#ssh $ip sudo systemctl stop xsuportal-api-golang
	#sudo tar zc /etc/systemd/system/xsuportal-*-golang.service | ssh tar zxv /
	rsync -av --delete /home/isucon/webapp/ $ip:/home/isucon/webapp/
	rsync -av --delete /home/isucon/env $ip:/home/isucon/env
	ssh $ip sudo systemctl start xsuportal-web-golang
	#ssh $ip sudo systemctl start xsuportal-api-golang
done

#sudo journalctl -u xsuportal-web-golang.service -f
sudo systemctl restart varnish
