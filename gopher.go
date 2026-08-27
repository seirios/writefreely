/*
 * Copyright © 2020 Musing Studio LLC.
 *
 * This file is part of WriteFreely.
 *
 * WriteFreely is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, included
 * in the LICENSE file in this source code package.
 */

package writefreely

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/writeas/web-core/log"
	"github.com/writefreely/go-gopher"
)

func initGopher(apper Apper) {
	handler := NewWFHandler(apper)

	gopher.HandleFunc("/", handler.Gopher(handleGopher))
	log.Info("Serving on gopher://localhost:%d", apper.App().Config().Server.GopherPort)
	gopher.ListenAndServe(fmt.Sprintf(":%d", apper.App().Config().Server.GopherPort), nil)
}

// Utility function to strip the URL from the hostname provided by app.cfg.App.Host
func stripHostProtocol(app *App) string {
	u, err := url.Parse(app.cfg.App.Host)
	if err != nil {
		// Fall back to host, with scheme stripped
		return string(regexp.MustCompile("^.*://").ReplaceAll([]byte(app.cfg.App.Host), []byte("")))
	}
	return u.Hostname()
}

func handleGopher(app *App, w gopher.ResponseWriter, r *gopher.Request) error {
	parts := strings.Split(r.Selector, "/")
	if parts[1] != "" {
		return handleGopherCollectionPost(app, w, r)
	}
	return handleGopherCollection(app, w, r)
}

func handleGopherCollection(app *App, w gopher.ResponseWriter, r *gopher.Request) error {
	var collAlias, slug string
	var c *Collection
	var err error
	var baseSel = "/"

	parts := strings.Split(r.Selector, "/")
	// sanity check
	slug = parts[1]
	if slug != "" {
		return handleGopherCollectionPost(app, w, r)
	}

	c, err = app.db.GetCollectionByID(1)
	if err != nil {
		return err
	}
	c.hostName = app.cfg.App.Host

	w.WriteInfo(c.DisplayTitle())
	if c.Description != "" {
		w.WriteInfo(c.Description)
	}

	posts, err := app.db.GetPosts(app.cfg, c, 0, false, false, false, "")
	if err != nil {
		return err
	}

	for _, p := range *posts {
		w.WriteItem(&gopher.Item{
			Port:        app.cfg.Server.GopherPort,
			Host:        stripHostProtocol(app),
			Type:        gopher.FILE,
			Description: p.CreatedDate() + " - " + p.DisplayTitle(),
			Selector:    baseSel + p.Slug.String,
		})
	}
	return w.End()
}

func handleGopherCollectionPost(app *App, w gopher.ResponseWriter, r *gopher.Request) error {
	var collAlias, slug string
	var c *Collection
	var err error

	parts := strings.Split(r.Selector, "/")
	slug = parts[1]
	c, err = app.db.GetCollectionByID(1)
	if err != nil {
		return err
	}
	c.hostName = app.cfg.App.Host

	p, err := app.db.GetPost(slug, c.ID)
	if err != nil {
		return err
	}

	b := bytes.Buffer{}
	if p.Title.String != "" {
		b.WriteString(p.Title.String + "\n")
	}
	b.WriteString(p.DisplayDate + "\n\n")
	b.WriteString(p.Content)
	io.Copy(w, &b)

	return w.End()
}
