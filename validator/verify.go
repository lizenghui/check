package validator

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Ver struct {
	proxy_url string
}

func NewVerify(url string) *Ver {

	return &Ver{url}
}

func requestURL(requrl string, proxy_url string, follow_redirect bool) (string, *http.Response) {
	proxy, _ := url.Parse("http://" + proxy_url)
	client := &http.Client{
		Timeout: 5 * time.Second,

		Transport: &http.Transport{
			// 设置代理
			Proxy: http.ProxyURL(proxy),
		},
	}

	if !follow_redirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	req, _ := http.NewRequest("GET", requrl, nil)

	req.Header.Set("USER-AGENT", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36")

	resp, err := client.Do(req)
	// resp, err := client.Get("http://myexternalip.com/raw")
	if err != nil {
		//return errors.New(strings.ReplaceAll(err.Error(), newrequrl, requrl))

		return "Error", resp
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	return strings.Trim(string(content), "\n"), resp
}

func (vs *Ver) Netflix() string {
	netflixUrl := "https://www.netflix.com/title/81280792"

	content, resp := requestURL(netflixUrl, vs.proxy_url, true)

	if content == "Error" {
		return "ERR"
	}

	if resp.StatusCode == 404 {
		return "Originals Only"
	} else if resp.StatusCode == 403 {
		return "N"
	} else if resp.StatusCode == 200 {
		return "Y"
	} else {
		// fmt.Println(resp)
		return "unknow"
	}

}

func (vs *Ver) CustomProbe(custom_url string) string {

	content, resp := requestURL(custom_url, vs.proxy_url, true)

	if content == "Error" {
		return "ERR"
	}

	if resp.StatusCode == 200 {
		return "Y"
	} else {
		// fmt.Println(resp)
		return "N"
	}

}
