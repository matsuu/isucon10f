#!/bin/sh

for ip in isu2.t.isucon.dev isu3.t.isucon.dev
do
	#ssh $ip sudo systemctl stop isuumo.go
	#rsync -av --delete /home/isucon/isuumo/webapp/go/ $ip:/home/isucon/isuumo/webapp/go/
	#ssh $ip sudo systemctl start isuumo.go
done
