#!/bin/bash

auto_code=$(grep "google:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort)

#echo $auto_code

cp -vf /home/lzh/etc/clash/config.yaml ./config.yaml.test

glados_auto_line=$(grep -n "\- name: ♻️ glados_Auto" config.yaml.test | cut -d ":" -f 1)
glados_auto_sed_line=$((glados_auto_line+6))

auto_line=$(grep -n "\- name: ♻️ auto2" config.yaml.test | cut -d ":" -f 1)
glados_auto_end_line=$((auto_line-1))

eval "$(sed $glados_auto_sed_line,${glados_auto_end_line}d config.yaml.test > config.yaml.1)"

IFS=$'\n'
for a in $auto_code
do
	eval "$(sed -i "${glados_auto_sed_line}i\\ $a" config.yaml.1) "
done
#echo "$glados_auto_sed_line" "$glados_auto_end_line"

