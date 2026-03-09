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
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/metacubex/mihomo/adapter/inbound"
	"github.com/metacubex/mihomo/component/nat"
	mconfig "github.com/metacubex/mihomo/config"
	"github.com/metacubex/mihomo/constant"
	_ "github.com/metacubex/mihomo/hub/executor"
	chttp "github.com/metacubex/mihomo/listener/http"
)

type switchingTunnel struct {
	natTable constant.NatTable

	mu    sync.RWMutex
	proxy constant.Proxy
}

func newSwitchingTunnel() *switchingTunnel {
	return &switchingTunnel{natTable: nat.New()}
}

func (t *switchingTunnel) SetProxy(p constant.Proxy) {
	t.mu.Lock()
	t.proxy = p
	t.mu.Unlock()
}

func (t *switchingTunnel) HandleTCPConn(conn net.Conn, metadata *constant.Metadata) {
	t.mu.RLock()
	p := t.proxy
	t.mu.RUnlock()

	if p == nil {
		_ = conn.Close()
		return
	}

	remote, err := p.DialContext(context.Background(), metadata)
	if err != nil {
		_ = conn.Close()
		return
	}

	relay(remote, conn)
}

func (t *switchingTunnel) HandleUDPPacket(packet constant.UDPPacket, _ *constant.Metadata) {
	if packet != nil {
		packet.Drop()
	}
}

func (t *switchingTunnel) NatTable() constant.NatTable {
	return t.natTable
}

type Args struct {
	config_path, port, ctype, custom_url string
}

var args = Args{}

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
	client := &http.Client{
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

	flag.StringVar(&args.config_path, "c", "config.yaml", "config file;")
	flag.StringVar(&args.port, "p", "18081", "proxy port;")
	flag.StringVar(&args.ctype, "t", "0", "check type; \n\t0:check netflix;\n\t1:check google&youtube premium US\n\t2:check chatGPT\n")
	flag.StringVar(&args.custom_url, "u", "", "custom probe url")

	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("\ngenerate mihomo config:\ngoogle&youtube:\n----------\ngrep \"youtube:Y\" 1.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort && echo \" \" && \\\ngrep \"google:Y\" 1.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort\n----------\nnetflix:\n----------\ngrep \"netflix:Y\" 0.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort\n----------\nchatGPT:\n----------\ngrep \"chatGPT:Y\" 2.check.log | cut -f 1 | cut -d \":\" -f 2 | sed 's/^/      -/g' | sort")
	}
	flag.Parse()

	proxy_url += args.port
	fmt.Println(proxy_url)

	log.SetFlags(0)

	fw, _ := os.OpenFile(args.ctype+".check.log", os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0666)
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

func parseMihomoConfigWithoutRules(path string) (map[string]constant.Proxy, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rawCfg, err := mconfig.UnmarshalRawConfig(buf)
	if err != nil {
		return nil, err
	}

	// This project only needs outbound proxies to dial probe requests.
	// Skip rules / rule-providers / GEOIP related parsing to avoid triggering GeoIP MMDB downloads.
	rawCfg.RuleProvider = nil
	rawCfg.Rule = nil
	rawCfg.SubRules = nil

	// DNS parsing may initialize GEOIP fallback filters in some configs; disable it here.
	rawCfg.DNS.Enable = false
	rawCfg.DNS.Fallback = nil
	rawCfg.DNS.FallbackFilter.GeoIP = false

	if len(rawCfg.DNS.DefaultNameserver) == 0 {
		rawCfg.DNS.DefaultNameserver = []string{"114.114.114.114"}
	}

	cfg, err := mconfig.ParseRawConfig(rawCfg)
	if err != nil {
		return nil, err
	}

	return cfg.Proxies, nil
}

func main() {
	_, err := os.Stat(args.config_path)

	if nil != err {
		fmt.Println("config illegal")
		fmt.Println(args.config_path)
		os.Exit(1)
	}

	nodes, err := parseMihomoConfigWithoutRules(args.config_path)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	// Allow local connections to the embedded HTTP proxy listener.
	// Mihomo's default HTTP listener applies inbound IP filters via adapter/inbound globals.
	inbound.SetAllowedIPs([]netip.Prefix{
		netip.MustParsePrefix("127.0.0.1/32"),
		netip.MustParsePrefix("::1/128"),
	})
	inbound.SetDisAllowedIPs(nil)

	tunnel := newSwitchingTunnel()

	l, err := chttp.New(proxy_url, tunnel)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	println("listen at:", l.Address())

	index := 1

	// total := len(nodes)

	for node, server := range nodes {

		var (
			res string
		)

		if server.Type() != constant.Shadowsocks && server.Type() != constant.ShadowsocksR && server.Type() != constant.Snell && server.Type() != constant.Socks5 && server.Type() != constant.Http && server.Type() != constant.Vmess && server.Type() != constant.Trojan {
			continue
		}
		tunnel.SetProxy(server)

		//落地机IP
		ip := getIpInfo()
		str := fmt.Sprintf("%d.node: %s", index, node)

		if args.custom_url != "" {
			vs := validator.NewVerify(proxy_url)
			res = "\t" + args.custom_url + ":" + vs.CustomProbe(args.custom_url)
		}

		re := regexp.MustCompile("美|波特兰|达拉斯|俄勒冈|凤凰城|费利蒙|硅谷|拉斯维加斯|洛杉矶|圣何塞|圣克拉拉|西雅图|芝加哥|US|United States")

		if args.ctype == "0" && !re.MatchString(node) {
			vs := validator.NewVerify(proxy_url)
			res = "\tnetflix:" + vs.Netflix()
		}

		if args.ctype == "1" {
			res += "\tgoogle:" + google()
			res += "\tyoutube:" + youtube_premium()

		}

		if args.ctype == "2" {
			vs := validator.NewVerify(proxy_url)
			res += "\tchatGPT:" + vs.ChatGPT()
		}

		logger.Printf("%s%s\t%s\n", str, res, ip)

		index++
	}
	defer fw.Close()
}
