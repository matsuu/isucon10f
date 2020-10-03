#!/bin/sh

for ip in isu3.t.isucon.dev
do
	ssh $ip sudo systemctl stop xsuportal-web-golang
	ssh $ip sudo systemctl stop xsuportal-api-golang
	#sudo tar zc /etc/systemd/system/xsuportal-*-golang.service | ssh tar zxv /
	rsync -av --delete /home/isucon/webapp/ $ip:/home/isucon/webapp/
	ssh $ip sudo systemctl start xsuportal-web-golang
	ssh $ip sudo systemctl start xsuportal-api-golang
done
