#!/bin/bash

cp -vf /home/lzh/etc/clash/config.yaml ./config.yaml.tmp

auto_code=$(grep "google:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort)
grep "google:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort
#echo $auto_code


glados_auto_line=$(grep -n "\- name: ♻️ glados_Auto" config.yaml.tmp | cut -d ":" -f 1)
glados_auto_sed_line=$((glados_auto_line+6))

auto_line=$(grep -n "\- name: ♻️ auto2" config.yaml.tmp | cut -d ":" -f 1)
glados_auto_end_line=$((auto_line-1))

eval "$(sed $glados_auto_sed_line,${glados_auto_end_line}d config.yaml.tmp > config.yaml.google)"

IFS=$'\n'
for a in $auto_code
do
	eval "$(sed -i "${glados_auto_sed_line}i\\ $a" config.yaml.google) "
done


grep "youtube:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort
auto_code=$(grep "youtube:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort)

glados_auto_line=$(grep -n "\- name: 🇺🇲 glados_US" config.yaml.google | cut -d ":" -f 1)
glados_auto_sed_line=$((glados_auto_line+6))

auto_line=$(grep -n "\- name: 🇺🇲 bywave_US" config.yaml.google | cut -d ":" -f 1)
glados_auto_end_line=$((auto_line-1))

eval "$(sed $glados_auto_sed_line,${glados_auto_end_line}d config.yaml.google > config.yaml.youtube)"

IFS=$'\n'
for a in $auto_code
do
	eval "$(sed -i "${glados_auto_sed_line}i\\ $a" config.yaml.youtube) "
done

#echo "$glados_auto_sed_line" "$glados_auto_end_line"

