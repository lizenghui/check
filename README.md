# check
self-use

check netflix
```
./check -c <mihomo_config> -t 0
```
check google & youtube & gemini
```
./check -c <mihomo_config> -t 1
```
check chatGPT
```
./check -c <mihomo_config> -t 2
```
check gemini
```
./check -c <mihomo_config> -t 3
```

generate mihomo config
```shell
grep "youtube:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort && echo " " && \
 grep "google:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort && echo " " && \
 grep "gemini:Y" 1.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort
 
grep "netflix:Y" 0.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort

grep "chatGPT:Y" 2.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort

grep "gemini:Y" 3.check.log | cut -f 1 | cut -d ":" -f 2 | sed 's/^/      -/g' | sort
```


## 感谢
1. [quzard/netflix-all-verify](https://github.com/quzard/netflix-all-verify)
2. [netflix-verify](https://github.com/sjlleo/netflix-verify)
3. [mihomo](https://github.com/MetaCubeX/mihomo)
