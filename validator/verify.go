package validator

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Ver struct {
	proxy_url string
}

const requestTimeout = 3 * time.Second

func NewVerify(url string) *Ver {

	return &Ver{url}
}

func requestURL(requrl string, proxy_url string, follow_redirect bool) (string, *http.Response) {
	proxy, _ := url.Parse("http://" + proxy_url)
	client := &http.Client{
		Timeout: requestTimeout,

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
		// fmt.Println(err)
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

func (vs *Ver) ChatGPT() string {
	gptUrl := "https://chat.openai.com/cdn-cgi/trace"

	re := regexp.MustCompile("AL|DZ|AD|AO|AG|AR|AM|AU|AT|AZ|BS|BD|BB|BE|BZ|BJ|BT|BO|BA|BW|BR|BN|BG|BF|CV|CA|CL|CO|KM|CG|CR|CI|HR|CY|CZ|DK|DJ|DM|DO|EC|SV|EE|FJ|FI|FR|GA|GM|GE|DE|GH|GR|GD|GT|GN|GW|GY|HT|VA|HN|HU|IS|IN|ID|IQ|IE|IL|IT|JM|JP|JO|KZ|KE|KI|KW|KG|LV|LB|LS|LR|LI|LT|LU|MG|MW|MY|MV|ML|MT|MH|MR|MU|MX|FM|MD|MC|MN|ME|MA|MZ|MM|NA|NR|NP|NL|NZ|NI|NE|NG|MK|NO|OM|PK|PW|PS|PA|PG|PY|PE|PH|PL|PT|QA|RO|RW|KN|LC|VC|WS|SM|ST|SN|RS|SC|SL|SG|SK|SI|SB|ZA|KR|ES|LK|SR|SE|CH|TW|TZ|TH|TL|TG|TO|TT|TN|TR|TV|UG|UA|AE|GB|US|UY|VU|ZM")

	content, _ := requestURL(gptUrl, vs.proxy_url, true)

	if content == "Error" {
		return "ERR"
	}

	contry_re := regexp.MustCompile(`loc=(\w+)`)
	contry := contry_re.FindStringSubmatch(content)

	if len(contry) > 1 && re.MatchString(contry[1]) {
		return "Y" + "(" + contry[1] + ")"
	} else {
		return "N"
	}

}

func geminiUnavailable(content string) bool {
	content = strings.ToLower(content)
	keywords := []string{
		"gemini isn't available in your country",
		"gemini is not available in your country",
		"gemini isn't available in your region",
		"gemini is not available in your region",
		"not available in your country",
		"not available in your region",
		"not supported in your country",
		"not supported in your region",
	}
	for _, k := range keywords {
		if strings.Contains(content, k) {
			return true
		}
	}
	return false
}

func (vs *Ver) Gemini() string {
	geminiUrl := "https://gemini.google.com/app"

	content, resp := requestURL(geminiUrl, vs.proxy_url, false)

	if content == "Error" || resp == nil {
		return "ERR"
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := strings.ToLower(resp.Header.Get("Location"))
		if strings.Contains(loc, "/sorry/") || strings.Contains(loc, "google.com/sorry") {
			return "N"
		}
		if strings.Contains(loc, "unsupported") || strings.Contains(loc, "unavailable") || strings.Contains(loc, "not-available") {
			return "N"
		}
		return "Y"
	}

	if resp.StatusCode == 403 || resp.StatusCode == 429 || resp.StatusCode == 451 {
		return "N"
	}

	if resp.StatusCode == 200 {
		if geminiUnavailable(content) {
			return "N"
		}
		return "Y"
	}

	if resp.StatusCode == 401 {
		return "Y"
	}

	return "N"
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
