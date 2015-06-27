/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mvdan/jutgelint"
)

const (
	// Name of the HTTP form fields when uploading code
	fieldCode = "code"
	fieldLang = "lang"

	// Content-Type when serving text
	contentType = "text/plain; charset=utf-8"

	// HTTP response strings
	invalidID     = "invalid id"
	unknownAction = "unsupported action"

	maxSize = 1 * 1024 * 1024
	timeout = 10 * time.Second

	cmdName = "jutge-lintd"
	version = "alpha1"
)

var (
	siteURL = flag.String("u", "http://localhost:8080", "URL of the site")
	listen  = flag.String("l", ":8080", "Host and port to listen to")
	workers = flag.Int("w", 4, "Number of POST workers")
)

func getCodeFromForm(r *http.Request) ([]byte, jutgelint.Lang, error) {
	var l jutgelint.Lang
	if code := r.FormValue(fieldCode); len(code) > 0 {
		lang, err := jutgelint.ParseLang(r.FormValue(fieldLang))
		if err != nil {
			return nil, l, err
		}
		return []byte(code), lang, nil
	}
	if f, h, err := r.FormFile(fieldCode); err == nil {
		defer f.Close()
		content, err := ioutil.ReadAll(f)
		if err == nil && len(content) > 0 {
			lang, err := jutgelint.ParseLangFilename(h.Filename)
			if err != nil {
				return nil, l, err
			}
			return content, lang, nil
		}
	}
	return nil, l, errors.New("no code provided")
}

func setHeaders(header http.Header, id ID, review Review) {
	modTime := review.ModTime()
	header.Set("Etag", fmt.Sprintf(`"%d-%s"`, modTime.Unix(), id))
	header.Set("Content-Type", contentType)
}

type httpHandler struct {
	store Store
	post  chan postRequest
}

type postRequest struct {
	code []byte
	lang jutgelint.Lang
	ret  chan postResult
}

type postResult struct {
	url string
	err error
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.handleGet(w, r)
	case "POST":
		h.handlePost(w, r)
	default:
		http.Error(w, unknownAction, http.StatusBadRequest)
	}
}

func logPrintfReq(r *http.Request, format string, v ...interface{}) {
	format = "%s %s %s " + format
	rv := []interface{}{r.RemoteAddr, r.Proto, r.Method}
	v = append(rv, v...)
	log.Printf(format, v...)
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if _, e := templates[r.URL.Path]; e {
		err := tmpl.ExecuteTemplate(w, r.URL.Path,
			struct {
				SiteURL   string
				FieldCode string
				FieldLang string
			}{
				SiteURL:   *siteURL,
				FieldCode: fieldCode,
				FieldLang: fieldLang,
			})
		if err != nil {
			logPrintfReq(r, "Error executing template for %s: %v", r.URL.Path, err)
		}
		return
	}
	id, err := IDFromString(r.URL.Path[1:])
	if err != nil {
		logPrintfReq(r, "Bad request: %v", err)
		http.Error(w, invalidID, http.StatusBadRequest)
		return
	}
	review, err := h.store.Get(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrReviewNotFound {
			status = http.StatusNotFound
		}
		logPrintfReq(r, "Error: %v", err)
		http.Error(w, err.Error(), status)
		return
	}
	defer review.Close()
	setHeaders(w.Header(), id, review)
	logPrintfReq(r, "Served: %s/%s", *siteURL, id)
	http.ServeContent(w, r, "", review.ModTime(), review)
}

func commentCode(code []byte, lang jutgelint.Lang) ([]byte, error) {
	in := bytes.NewReader(code)
	var out bytes.Buffer
	if err := jutgelint.CheckAndCommentCode(lang, in, &out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (h *httpHandler) postWorker() {
	for {
		req := <-h.post
		comm, err := commentCode(req.code, req.lang)
		if err != nil {
			req.ret <- postResult{err: err}
			continue
		}
		id, err := h.store.Put(comm)
		if err != nil {
			req.ret <- postResult{err: err}
			continue
		}
		req.ret <- postResult{
			url: fmt.Sprintf("%s/%s", *siteURL, id),
		}
	}
}

func (h *httpHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	code, lang, err := getCodeFromForm(r)
	if err != nil {
		logPrintfReq(r, "Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret := make(chan postResult)
	h.post <- postRequest{
		code: code,
		lang: lang,
		ret:  ret,
	}
	res := <-ret
	if res.err != nil {
		logPrintfReq(r, "Error: %v", res.err)
		http.Error(w, res.err.Error(), http.StatusInternalServerError)
		return
	}
	logPrintfReq(r, "Created: %s", res.url)
	if r.URL.Path == "/redirect" {
		http.Redirect(w, r, res.url, 302)
	} else {
		fmt.Fprintln(w, res.url)
	}
}

func main() {
	flag.Parse()
	loadTemplates()
	if *workers < 1 {
		log.Fatalf("Cannot have less than 1 workers!")
	}
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("%s %s ", cmdName, version))
	log.Printf("siteURL = %s", *siteURL)
	log.Printf("listen  = %s", *listen)
	log.Printf("workers = %d", *workers)

	handler := httpHandler{
		post: make(chan postRequest),
	}
	var err error
	handler.store, err = NewFileStore("code")
	if err != nil {
		log.Fatalf("Could not setup file store: %v", err)
	}

	for i := 0; i < *workers; i++ {
		go handler.postWorker()
	}
	//http.Handle("/", http.TimeoutHandler(handler, timeout, ""))
	http.Handle("/", handler)
	log.Println("Up and running!")
	log.Fatal(http.ListenAndServe(*listen, nil))
}
