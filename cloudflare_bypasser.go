package bypasser

import (
    "encoding/base64"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os/exec"
    "regexp"
)

type Client struct {
    session *http.Client
}

func parseParams(html string) (map[string]string, error) {
    var (
        re     = regexp.MustCompile(`name="(s|jschl_vc|pass)"(?: [^<>]*)? value="(.+?)"`)
        caps   = re.FindAllStringSubmatch(html, 3)
        params = make(map[string]string, 4)
    )

    if len(caps) == 0 {
        return nil, errors.New("regexp failed, in `func parseParams(html string) (map[string]string, error)`")
    } else {
        for _, param := range caps {
            params[param[1]] = param[2]
        }

        return params, nil
    }
}

func parseChallenge(html string, domain string) string {
    var (
        re           = regexp.MustCompile(`(?s)setTimeout\(function\(\)\{\s+(var s,t,o,p,b,r,e,a,k,i,n,g,f.+?\r?\n[\s\S]+?a\.value =.+?)\r?\n`)
        challenge    = re.FindStringSubmatch(html)[1]
        b64Challenge = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
            `
			var document = {
                createElement: function () {
                    return { firstChild: { href: "http://%s/" } }
                },
                getElementById: function () {
                    return {"innerHTML": ""};
                }
            };
            %s; a.value`,
            domain,
            challenge,
        )))
    )

    return fmt.Sprintf(
        `
		var atob = Object.setPrototypeOf(function (str) {
            try {
                return Buffer.from("" + str, "base64").toString("binary");
            } catch (e) {}
        }, null);
        var challenge = atob("%s");
        var context = Object.setPrototypeOf({ atob: atob }, null);
        var options = {
            filename: "iuam-challenge.js",
            contextOrigin: "cloudflare:iuam-challenge.js",
            contextCodeGeneration: { strings: true, wasm: false },
            timeout: 5000
        };
        process.stdout.write(String(
            require("vm").runInNewContext(challenge, context, options)
        ));`,
        b64Challenge,
    )
}

func runJs(js string) (string, error) {
    var cmd = exec.Command("node", "-e", js)

    if output, e := cmd.Output(); e == nil {
        return string(output), nil
    } else {
        return "", e
    }
}

func NewBypasser(session *http.Client) Client {
    if session.Jar == nil {
        session.Jar, _ = cookiejar.New(nil)
    }

    return Client{session: session}
}

func (bypasser *Client) Solve(r *http.Request, retry uint) (*url.URL, []*http.Cookie, error) {
    var (
        userAgent = r.Header.Get("User-Agent")
        u         = r.URL

        e    error
        resp *http.Response

        content []byte
        html    string

        params      map[string]string
        domain      string
        challenge   string
        jschlAnswer string
        query       url.Values
    )

    if userAgent == "" {
        userAgent = "Mozilla/5.0"
        r.Header.Set("User-Agent", userAgent)
    }

    resp, e = bypasser.session.Do(r)
    defer resp.Body.Close()
    if e != nil {
        return nil, nil, e
    }

    content, e = ioutil.ReadAll(resp.Body)
    if e != nil {
        return nil, nil, e
    }
    html = string(content)

    params, e = parseParams(html)
    if e != nil {
        return nil, nil, e
    }

    domain = u.Host
    challenge = parseChallenge(html, domain)
    jschlAnswer, e = runJs(challenge)
    if e != nil {
        return nil, nil, e
    }
    params["jschl_answer"] = jschlAnswer

    r, _ = http.NewRequest("GET", fmt.Sprintf("%s://%s/cdn-cgi/l/chk_jschl", u.Scheme, domain), nil)
    r.Header.Set("Referer", u.String())
    r.Header.Set("User-Agent", userAgent)
    query = r.URL.Query()
    for k, v := range params {
        query.Add(k, v)
    }
    r.URL.RawQuery = query.Encode()

    for i := uint(1); i != retry; i += 1 {
        resp, e = bypasser.session.Do(r)
        defer resp.Body.Close()
        if e != nil {
            return nil, nil, e
        }

        if resp.StatusCode == 200 {
            break
        }
    }

    return u, bypasser.session.Jar.Cookies(u), nil
}
