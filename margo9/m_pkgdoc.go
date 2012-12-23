package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type mPkgdoc struct {
	Q    jString
	Path jString
}

type mPkgdocDoc struct {
	Path string
	Doc  string
}

func mPkgdocFetchDoc(m *mPkgdoc) (interface{}, string) {
	res := M{}
	path := strings.TrimSpace(m.Path.String())
	if path == "" {
		return res, "invalid query"
	}

	req, err := http.NewRequest("GET", "http://godoc.org/"+path, nil)
	if err != nil {
		return res, errStr(err)
	}

	req.Header.Add("Accept", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, errStr(err)
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return res, errStr(err)
	}

	res["doc"] = mPkgdocDoc{
		Path: path,
		Doc:  string(s),
	}
	return res, ""
}

func mPkgdocSearch(m *mPkgdoc) (interface{}, string) {
	res := M{}
	s := strings.TrimSpace(m.Q.String())
	if s == "" {
		return res, "invalid query"
	}

	req, err := http.NewRequest("GET", "http://godoc.org/?q="+url.QueryEscape(s), nil)
	if err != nil {
		return res, errStr(err)
	}

	req.Header.Set("User-Agent", "GoSublime")
	req.Header.Add("Accept", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, errStr(err)
	}
	defer resp.Body.Close()

	packages := []mPkgdocDoc{}
	rd := bufio.NewReader(resp.Body)
	for {
		s, err := rd.ReadBytes('\n')
		if err != nil {
			break
		}

		s = bytes.TrimSpace(s)
		if len(s) > 0 {
			l := bytes.SplitN(s, []byte{' '}, 2)
			v := mPkgdocDoc{
				Path: string(l[0]),
			}

			if len(l) == 2 {
				v.Doc = string(l[1])
			}
			packages = append(packages, v)
		}
	}

	res["packages"] = packages
	return res, ""
}

func (m *mPkgdoc) Call() (interface{}, string) {
	if m.Q != "" {
		return mPkgdocSearch(m)
	}
	return mPkgdocFetchDoc(m)
}

func init() {
	registry.Register("pkgdoc", func(b *Broker) Caller {
		return &mPkgdoc{}
	})
}
