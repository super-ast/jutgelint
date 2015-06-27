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
)

var (
	siteURL = flag.String("u", "http://localhost:8080", "URL of the site")
	listen  = flag.String("l", ":8080", "Host and port to listen to")
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

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if _, e := templates[r.URL.Path]; e {
		err := tmpl.ExecuteTemplate(w, r.URL.Path,
			struct {
				SiteURL   string
				FieldCode string
			}{
				SiteURL:   *siteURL,
				FieldCode: fieldCode,
			})
		if err != nil {
			log.Printf("Error executing template for %s: %v", r.URL.Path, err)
		}
		return
	}
	id, err := IDFromString(r.URL.Path[1:])
	if err != nil {
		http.Error(w, invalidID, http.StatusBadRequest)
		return
	}
	review, err := h.store.Get(id)
	if err == ErrReviewNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Unknown error on GET: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer review.Close()
	setHeaders(w.Header(), id, review)
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

func (h *httpHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	code, lang, err := getCodeFromForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	comm, err := commentCode(code, lang)
	if err != nil {
		log.Printf("Could not check and comment code: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := h.store.Put(comm)
	if err != nil {
		log.Printf("Unknown error on POST: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s/%s\n", *siteURL, id)
}

func main() {
	flag.Parse()
	loadTemplates()
	log.Printf("siteURL    = %s", *siteURL)
	log.Printf("listen     = %s", *listen)

	var handler httpHandler
	var err error
	handler.store, err = NewFileStore("code")
	if err != nil {
		log.Fatalf("Could not setup file store: %v", err)
	}

	finalHandler := http.TimeoutHandler(handler, timeout, "")
	http.Handle("/", finalHandler)
	log.Println("Up and running!")
	log.Fatal(http.ListenAndServe(*listen, nil))
}
