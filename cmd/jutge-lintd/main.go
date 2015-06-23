/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	// Name of the HTTP form field when uploading code
	fieldName = "code"
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

func getContentFromForm(r *http.Request) ([]byte, error) {
	if value := r.FormValue(fieldName); len(value) > 0 {
		return []byte(value), nil
	}
	if f, _, err := r.FormFile(fieldName); err == nil {
		defer f.Close()
		content, err := ioutil.ReadAll(f)
		if err == nil && len(content) > 0 {
			return content, nil
		}
	}
	return nil, errors.New("no code provided")
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
				FieldName string
			}{
				SiteURL:   *siteURL,
				FieldName: fieldName,
			})
		if err != nil {
			log.Printf("Error executing template for %s: %s", r.URL.Path, err)
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
		log.Printf("Unknown error on GET: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer review.Close()
	setHeaders(w.Header(), id, review)
	http.ServeContent(w, r, "", review.ModTime(), review)
}

func (h *httpHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	content, err := getContentFromForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.store.Put(content)
	if err != nil {
		log.Printf("Unknown error on POST: %s", err)
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
		log.Fatalf("Could not setup file store: %s", err)
	}

	finalHandler := http.TimeoutHandler(handler, timeout, "")
	http.Handle("/", finalHandler)
	log.Println("Up and running!")
	log.Fatal(http.ListenAndServe(*listen, nil))
}
