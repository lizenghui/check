package main

import (
	"check/validator"
	"context"
	"flag"
	"fmt"
	"io"

	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub/executor"
	chttp "github.com/Dreamacro/clash/listener/http"
)

var proxy constant.Proxy
var config_path string
var port string
var ctype string
var proxy_url = "127.0.0.1:"
var fw os.File
var logger *log.Logger

func relay(l, r net.Conn) {
	go io.Copy(l, r)
	io.Copy(r, l)
}

func getIpInfo() string {
	body := requestURL("http://myip.ipip.net")
	if len(body) > 100 {
		return strings.Replace(body[0:50], "\n", " ", -1)
	} else {
		return body
	}
}

func requestURL(requrl string) string {
	proxy, _ := url.Parse("http://" + proxy_url)
	client := http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{
			// 设置代理
			Proxy: http.ProxyURL(proxy),
		},
	}

	req, _ := http.NewRequest("GET", requrl, nil)

	req.Header.Set("USER-AGENT", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36")

	resp, err := client.Do(req)
	// resp, err := client.Get("http://myexternalip.com/raw")
	if err != nil {
		//return errors.New(strings.ReplaceAll(err.Error(), newrequrl, requrl))
		return "Error"
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	return strings.Trim(string(content), "\n")
}

func init() {

	flag.StringVar(&config_path, "c", "config.yaml", "config file;")
	flag.StringVar(&port, "p", "18081", "proxy port;")
	flag.StringVar(&ctype, "t", "0", "check type; \n\t0:check netflix;\n\t1:check google&youtube premium US\n")

	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("\ngenerate clash config:\ngoogle&youtube:\n----------\ngrep \"youtube:Y\" 1.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort && echo \" \" && \\\ngrep \"google:Y\" 1.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort\n----------\nnetflix:\n----------\ngrep \"netflix:Y\" 0.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort")
	}
	flag.Parse()

	proxy_url += port
	fmt.Println(proxy_url)

	log.SetFlags(0)

	fw, _ := os.OpenFile(ctype+".check.log", os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0666)
	logger = log.New(io.MultiWriter(os.Stdout, fw), "", 0)
}

func youtube_premium() string {
	youtubeUrl := "https://www.youtube.com/premium"

	content := requestURL(youtubeUrl)
	if content == "Error" {
		return "ERR"
	}
	is := strings.Contains(content, "Premium is not available in your country")
	if is {
		//存在
		//fmt.Println(content)
		return "N"
	} else {
		//不存在
		return "Y"
	}
}

func google() string {
	googleUrl := "https://www.google.com"

	content := requestURL(googleUrl)
	if content == "Error" {
		return "ERR"
	}
	// fmt.Println(content)
	is := strings.Contains(content, "302 Moved")
	if is {
		//存在
		//fmt.Println(content)
		return "N"
	} else {
		//不存在
		return "Y"
	}
}

func main() {
	_, err := os.Stat(config_path)

	if nil != err {
		fmt.Println("config illegal")
		fmt.Println(config_path)
		os.Exit(1)
	}

	config, err := executor.ParseWithPath(config_path)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	in := make(chan constant.ConnContext, 100)
	defer close(in)
	l, err := chttp.New(proxy_url, in)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	println("listen at:", l.Address())

	go func() {
		for c := range in {
			conn := c
			metadata := conn.Metadata()
			go func() {
				remote, err := proxy.DialContext(context.Background(), metadata)

				if err != nil {
					conn.Conn().Close()
					// fmt.Println(err.Error())
					return
				}
				relay(remote, conn.Conn())
			}()
		}
	}()

	index := 1
	nodes := config.Proxies

	// total := len(nodes)

	for node, server := range nodes {

		var (
			res string
		)

		if server.Type() != constant.Shadowsocks && server.Type() != constant.ShadowsocksR && server.Type() != constant.Snell && server.Type() != constant.Socks5 && server.Type() != constant.Http && server.Type() != constant.Vmess && server.Type() != constant.Trojan {
			continue
		}
		proxy = server

		//落地机IP
		ip := getIpInfo()
		str := fmt.Sprintf("%d.node: %s", index, node)

		re := regexp.MustCompile("美|波特兰|达拉斯|俄勒冈|凤凰城|费利蒙|硅谷|拉斯维加斯|洛杉矶|圣何塞|圣克拉拉|西雅图|芝加哥|US|United States")

		if re.MatchString(node) {
			res += "\tyoutube:" + youtube_premium()
		} else {
			res += "\t"
		}

		if ctype == "0" && !re.MatchString(node) {
			vs := validator.NewVerify(proxy_url)
			res = "\tnetflix:" + vs.Netflix()
		} else {
			continue
		}

		if ctype == "1" {

			res += "\tgoogle:" + google()

			if re.MatchString(node) {
				res += "\tyoutube:" + youtube_premium()
			} else {
				res += "\t"
			}
		}
		logger.Printf("%s%s\t%s\n", str, res, ip)

		index++
	}
	defer fw.Close()
}
