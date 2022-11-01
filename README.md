# check
self-use

check netflix
```
./check -c <clash_config> -t 0
```
check google & youtube
```
./check -c <clash_config> -t 1
```

generate clash config
```shell
grep "youtube:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort && echo " " && grep "google:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort
grep "netflix:Y" 0.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort
```


## 感谢
1. [quzard/netflix-all-verify](https://github.com/quzard/netflix-all-verify)
2. [netflix-verify](https://github.com/sjlleo/netflix-verify)
3. [clash](https://github.com/Dreamacro/clash)