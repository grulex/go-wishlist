package scrapper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	escapedFragment string = "_escaped_fragment_="
	fragmentRegexp         = regexp.MustCompile("#!(.*)")
)

type scraper struct {
	Url                *url.URL
	EscapedFragmentUrl *url.URL
	MaxRedirect        int
}

type Document struct {
	Body    bytes.Buffer
	Preview documentPreview
}

type documentPreview struct {
	Icon        string
	Name        string
	Title       string
	Description string
	Images      []string
	Link        string
}

func Scrape(uri string, maxRedirect int) (*Document, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return (&scraper{Url: u, MaxRedirect: maxRedirect}).Scrape()
}

func (scraper *scraper) Scrape() (*Document, error) {
	doc, err := scraper.getDocument()
	if err != nil {
		return nil, err
	}
	err = scraper.parseDocument(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (scraper *scraper) getUrl() string {
	if scraper.EscapedFragmentUrl != nil {
		return scraper.EscapedFragmentUrl.String()
	}
	return scraper.Url.String()
}

func (scraper *scraper) toFragmentUrl() error {
	unescapedurl, err := url.QueryUnescape(scraper.Url.String())
	if err != nil {
		return err
	}
	matches := fragmentRegexp.FindStringSubmatch(unescapedurl)
	if len(matches) > 1 {
		escapedFragment := escapedFragment
		for _, r := range matches[1] {
			b := byte(r)
			if avoidByte(b) {
				continue
			}
			if escapeByte(b) {
				escapedFragment += url.QueryEscape(string(r))
			} else {
				escapedFragment += string(r)
			}
		}

		p := "?"
		if len(scraper.Url.Query()) > 0 {
			p = "&"
		}
		fragmentUrl, err := url.Parse(strings.Replace(unescapedurl, matches[0], p+escapedFragment, 1))
		if err != nil {
			return err
		}
		scraper.EscapedFragmentUrl = fragmentUrl
	} else {
		p := "?"
		if len(scraper.Url.Query()) > 0 {
			p = "&"
		}
		fragmentUrl, err := url.Parse(unescapedurl + p + escapedFragment)
		if err != nil {
			return err
		}
		scraper.EscapedFragmentUrl = fragmentUrl
	}
	return nil
}

func (scraper *scraper) getDocument() (*Document, error) {
	scraper.MaxRedirect -= 1
	if strings.Contains(scraper.Url.String(), "#!") {
		_ = scraper.toFragmentUrl()
	}
	if strings.Contains(scraper.Url.String(), escapedFragment) {
		scraper.EscapedFragmentUrl = scraper.Url
	}

	req, err := http.NewRequest("GET", scraper.getUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.2 Safari/605.1.15")

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	if err != nil {
		return nil, err
	}

	if resp.Request.URL.String() != scraper.getUrl() {
		scraper.EscapedFragmentUrl = nil
		scraper.Url = resp.Request.URL
	}
	b, err := convertUTF8(resp.Body, resp.Header.Get("content-type"))
	if err != nil {
		return nil, err
	}
	doc := &Document{Body: b, Preview: documentPreview{Link: scraper.Url.String()}}

	return doc, nil
}

func convertUTF8(content io.Reader, contentType string) (bytes.Buffer, error) {
	buff := bytes.Buffer{}
	content, err := charset.NewReader(content, contentType)
	if err != nil {
		return buff, err
	}
	_, err = io.Copy(&buff, content)
	if err != nil {
		return buff, err
	}
	return buff, nil
}

func (scraper *scraper) parseDocument(doc *Document) error {
	t := html.NewTokenizer(&doc.Body)
	var ogImage bool
	var headPassed bool
	var hasFragment bool
	var hasCanonical bool
	var canonicalUrl *url.URL
	doc.Preview.Images = []string{}
	// saves previews' link in case that <link rel="canonical"> is found after <meta property="og:url">
	link := doc.Preview.Link
	// set default value to site name if <meta property="og:site_name"> not found
	doc.Preview.Name = scraper.Url.Host
	// set default icon to web root if <link rel="icon" href="/favicon.ico"> not found
	doc.Preview.Icon = fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, "/favicon.ico")
	for {
		tokenType := t.Next()
		if tokenType == html.ErrorToken {
			return nil
		}
		if tokenType != html.SelfClosingTagToken && tokenType != html.StartTagToken && tokenType != html.EndTagToken {
			continue
		}
		token := t.Token()

		switch token.Data {
		case "head":
			if tokenType == html.EndTagToken {
				headPassed = true
			}
		case "body":
			headPassed = true

		case "link":
			var canonical bool
			var hasIcon bool
			var href string
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "rel" && cleanStr(attr.Val) == "canonical" {
					canonical = true
				}
				if cleanStr(attr.Key) == "rel" && strings.Contains(cleanStr(attr.Val), "icon") {
					hasIcon = true
				}
				if cleanStr(attr.Key) == "href" {
					href = attr.Val
				}
				if len(href) > 0 && canonical && link != href {
					hasCanonical = true
					var err error
					canonicalUrl, err = url.Parse(href)
					if err != nil {
						return err
					}
				}
				if len(href) > 0 && hasIcon {
					doc.Preview.Icon = href
				}
			}

		case "meta":
			if len(token.Attr) != 2 {
				break
			}
			if metaFragment(token) && scraper.EscapedFragmentUrl == nil {
				hasFragment = true
			}
			var property string
			var content string
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "property" || cleanStr(attr.Key) == "name" {
					property = attr.Val
				}
				if cleanStr(attr.Key) == "content" {
					content = attr.Val
				}
			}
			switch cleanStr(property) {
			case "twitter:title":
				doc.Preview.Title = content
			case "twitter:url":
				doc.Preview.Link = content
			case "twitter:description":
				doc.Preview.Description = content
			case "og:title":
				if len(doc.Preview.Title) == 0 {
					doc.Preview.Title = content
				}
			case "og:description":
				if len(doc.Preview.Description) == 0 {
					doc.Preview.Description = content
				}
			case "description":
				if len(doc.Preview.Description) == 0 {
					doc.Preview.Description = content
				}
			case "og:url":
				if len(doc.Preview.Link) == 0 {
					doc.Preview.Link = content
				}
			case "og:image":
				ogImage = true
				ogImgUrl, err := url.Parse(content)
				if err != nil {
					return err
				}
				if !ogImgUrl.IsAbs() {
					ogImgUrl, err = url.Parse(fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, ogImgUrl.Path))
					if err != nil {
						return err
					}
				}

				doc.Preview.Images = []string{ogImgUrl.String()}

			}

		case "title":
			if tokenType == html.StartTagToken {
				t.Next()
				token = t.Token()
				if len(doc.Preview.Title) == 0 {
					doc.Preview.Title = token.Data
				}
			}

		case "img":
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "src" {
					imgUrl, err := url.Parse(attr.Val)
					if err != nil {
						return err
					}
					if !imgUrl.IsAbs() {
						doc.Preview.Images = append(doc.Preview.Images, fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, imgUrl.Path))
					} else {
						doc.Preview.Images = append(doc.Preview.Images, attr.Val)
					}

				}
			}
		}

		if hasCanonical && headPassed && scraper.MaxRedirect > 0 {
			if !canonicalUrl.IsAbs() {
				absCanonical, err := url.Parse(fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, canonicalUrl.Path))
				if err != nil {
					return err
				}
				canonicalUrl = absCanonical
			}
			scraper.Url = canonicalUrl
			scraper.EscapedFragmentUrl = nil
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}

		if hasFragment && headPassed && scraper.MaxRedirect > 0 {
			_ = scraper.toFragmentUrl()
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}

		if len(doc.Preview.Title) > 0 && len(doc.Preview.Description) > 0 && ogImage && headPassed {
			return nil
		}

	}
}

func avoidByte(b byte) bool {
	i := int(b)
	if i == 127 || (i >= 0 && i <= 31) {
		return true
	}
	return false
}

func escapeByte(b byte) bool {
	i := int(b)
	if i == 32 || i == 35 || i == 37 || i == 38 || i == 43 || (i >= 127 && i <= 255) {
		return true
	}
	return false
}

func metaFragment(token html.Token) bool {
	var name string
	var content string

	for _, attr := range token.Attr {
		if cleanStr(attr.Key) == "name" {
			name = attr.Val
		}
		if cleanStr(attr.Key) == "content" {
			content = attr.Val
		}
	}
	if name == "fragment" && content == "!" {
		return true
	}
	return false
}

func cleanStr(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}
