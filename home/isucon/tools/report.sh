#!/bin/sh

now=`date +%H%M%S`
report_dir="${HOME}/report"
output_dir="${report_dir}/${now}"
latest_dir="${report_dir}/latest"

mkdir -p "${output_dir}"

echo "kataribe"
sudo cat /var/log/nginx/access.log | "${HOME}/tools/kataribe" > "${output_dir}/slow-kataribe.txt"

echo "pt-query-digest"
sudo cat /var/log/mysql/mysql-slow.log | "${HOME}/tools/pt-query-digest" --limit 100% > "${output_dir}/slow-mysql.txt"

echo "mysqltuner"
sudo perl ${HOME}/tools/mysqltuner.pl > "${output_dir}/mysqltuner.txt"

echo "error logs"
sudo cat /var/log/nginx/error.log > "${output_dir}/error-nginx.txt"
sudo cat /var/log/mysql/error.log > "${output_dir}/error-mysql.txt"

#for ip in s2 s3 ; do
#  echo "$ip"
#  ssh $ip sudo cat /var/log/mysql/mysql-slow.log | "${HOME}/tools/pt-query-digest" --limit 100% > "${output_dir}/slow-mysql-${ip}.txt"
#  ssh $ip sudo perl ${HOME}/tools/mysqltuner.pl > "${output_dir}/mysqltuner-${ip}.txt"
#  ssh $ip sudo cat /var/log/mysql/error.log > "${output_dir}/error-mysql-${ip}.txt"
#done

rm -f "${latest_dir}"
ln -s ${now} "${latest_dir}"
