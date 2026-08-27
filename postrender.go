/*
 * Copyright © 2018-2021 Musing Studio LLC.
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
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	stripmd "github.com/writeas/go-strip-markdown/v2"
	"github.com/writeas/impart"
	"github.com/gomarkdown/markdown"
	mdAST "github.com/gomarkdown/markdown/ast"
	mdHTML "github.com/gomarkdown/markdown/html"
	mdParser "github.com/gomarkdown/markdown/parser"
	"github.com/writeas/web-core/log"
	"github.com/writeas/web-core/stringmanip"
	"github.com/writefreely/writefreely/config"
	"github.com/writefreely/writefreely/parse"
)

var (
	blockReg        = regexp.MustCompile("<(ul|ol|blockquote)>\n")
	endBlockReg     = regexp.MustCompile("</([a-z]+)>\n</(ul|ol|blockquote)>")
	youtubeReg      = regexp.MustCompile("(https?://www.youtube.com/embed/[a-zA-Z0-9\\-_]+)(\\?[^\t\n\f\r \"']+)?")
	titleElementReg = regexp.MustCompile("</?h[1-6]>")
	hashtagReg      = regexp.MustCompile(`^#(([\d]+[\p{L}\p{M}]+[\p{L}\p{M}\d]*)|([\p{L}\p{M}][\p{L}\p{M}\d]*))`)
	mentionReg      = regexp.MustCompile(`^@([A-Za-z0-9._%+-]+)(@[A-Za-z0-9.-]+\.[A-Za-z]+)\b`)
	paragraphReg    = regexp.MustCompile("<p>(.+)</p>")
)

func (p *Post) formatContent(cfg *config.Config, c *Collection, isOwner bool, isPostPage bool) {
	baseURL := c.CanonicalURL()

	p.HTMLTitle = template.HTML(applyBasicMarkdown([]byte(p.Title.String)))
	p.HTMLContent = template.HTML(applyMarkdown([]byte(p.Content), baseURL, cfg))
	if exc := strings.Index(string(p.Content), "<!--more-->"); exc > -1 {
		p.HTMLExcerpt = template.HTML(applyMarkdown([]byte(p.Content[:exc]), baseURL, cfg))
	}
}

func (p *PublicPost) formatContent(cfg *config.Config, isOwner bool, isPostPage bool) {
	p.Post.formatContent(cfg, &p.Collection.Collection, isOwner, isPostPage)
}

func (p *Post) augmentContent(c *Collection) {
	if p.PinnedPosition.Valid {
		// Don't augment posts that are pinned
		return
	}
	if strings.Index(p.Content, shortCodeNoSig) > -1 {
		// Don't augment posts with the special "nosig" shortcode
		return
	}
	// Add post signatures
	if c.Signature != "" && !strings.HasSuffix(strings.TrimSpace(p.Content), strings.TrimSpace(c.Signature)) {
		p.Content += "\n\n" + c.Signature
	}
}

func (p *PublicPost) augmentContent() {
	p.Post.augmentContent(&p.Collection.Collection)
}

func applyMarkdown(data []byte, baseURL string, cfg *config.Config) string {
	return applyMarkdownSpecial(data, baseURL, cfg)
}

func disableYoutubeAutoplay(outHTML string) string {
	for _, match := range youtubeReg.FindAllString(outHTML, -1) {
		u, err := url.Parse(match)
		if err != nil {
			continue
		}
		u.RawQuery = html.UnescapeString(u.RawQuery)
		q := u.Query()
		// Set Youtube autoplay url parameter, if any, to 0
		if len(q["autoplay"]) == 1 {
			q.Set("autoplay", "0")
		}
		u.RawQuery = q.Encode()
		cleanURL := u.String()
		outHTML = strings.Replace(outHTML, match, cleanURL, 1)
	}
	return outHTML
}

func applyMarkdownSpecial(data []byte, baseURL string, cfg *config.Config) string {

	mdExtensions := 0 |
		mdParser.Tables |
		mdParser.FencedCode |
		mdParser.Autolink |
		mdParser.Strikethrough |
		mdParser.SpaceHeadings |
		mdParser.AutoHeadingIDs |
		mdParser.DefinitionLists
	htmlFlags := 0 |
		mdHTML.Smartypants |
		mdHTML.SmartypantsDashes

	p := mdParser.NewWithExtensions(mdExtensions)

	/* Hashtag parsing */
	var processHashtag = func(_ *mdParser.Parser, data []byte, offset int) (int, mdAST.Node) {
		data = data[offset:]

		var match []int = hashtagReg.FindIndex(data)
		// failed regex match => not a tag
		if match == nil {
			return 0, nil
		}
		data = data[match[0] + 1:match[1]]
		nb := match[1] - match[0]

		tagPrefix := baseURL + "tag:"
		// append link
		link := &mdAST.Link{
			Destination: append([]byte(tagPrefix), data...),
			AdditionalAttributes: []string{`class="hashtag"`},
		}
		mdAST.AppendChild(link, &mdAST.HTMLSpan{
			Leaf: mdAST.Leaf{
				Literal: append(append([]byte(`<span>#</span><span class="p-category">`), data...), []byte(`</span>`)...),
			}})
		return nb, link
	}

	/* Mention processing */
	var processMention = func(_ *mdParser.Parser, data []byte, offset int) (int, mdAST.Node) {
		data = data[offset:]

		var match []int = mentionReg.FindIndex(data)
		// failed regex match => not a mention
		if match == nil {
			return 0, nil
		}
		data = data[match[0] + 1:match[1]]
		nb := match[1] - match[0]

		handlePrefix := cfg.App.Host + "/@/"
		// append link
		link := &mdAST.Link{
			Destination: append([]byte(handlePrefix), data...),
			AdditionalAttributes: []string{`class="u-url mention"`},
		}
		mdAST.AppendChild(link, &mdAST.HTMLSpan{
			Leaf: mdAST.Leaf{
				Literal: append(append([]byte(`@<span>`), data...), []byte(`</span>`)...),
			}})
		return nb, link
	}

	if baseURL != "" {
		p.RegisterInline('#', processHashtag)
		p.RegisterInline('@', processMention)
	}

	doc := p.Parse([]byte(data))
	opts := mdHTML.RendererOptions{Flags: htmlFlags}
	renderer := mdHTML.NewRenderer(opts)

	// Generate Markdown
	md := markdown.Render(doc, renderer)

	// Strip out bad HTML
	policy := getSanitizationPolicy()
	policy.RequireNoFollowOnLinks(false)
	outHTML := string(policy.SanitizeBytes(md))
	// Strip newlines on certain block elements that render with them
	outHTML = blockReg.ReplaceAllString(outHTML, "<$1>")
	outHTML = endBlockReg.ReplaceAllString(outHTML, "</$1></$2>")
	outHTML = disableYoutubeAutoplay(outHTML)
	// Unescape entities
	outHTML = html.UnescapeString(outHTML)

	return outHTML
}

func applyBasicMarkdown(data []byte) string {
	if len(bytes.TrimSpace(data)) == 0 {
		return ""
	}

	mdExtensions := 0 |
		mdParser.Strikethrough |
		mdParser.SpaceHeadings
	htmlFlags := 0 |
		mdHTML.SkipHTML |
		mdHTML.Smartypants |
		mdHTML.SmartypantsDashes

	// Generate Markdown
	// This passes the supplied title into markdown.Render() as an H1 header, so we only render HTML that
	// belongs in an H1.
	heading := append([]byte("# "), data...)
	p := mdParser.NewWithExtensions(mdExtensions)
	doc := p.Parse(heading)
	opts := mdHTML.RendererOptions{Flags: htmlFlags}
	renderer := mdHTML.NewRenderer(opts)
	md := markdown.Render(doc, renderer)
	// Remove H1 markup
	md = bytes.TrimSpace(md) // markdown.Render adds a newline at the end of the <h1>
	if len(md) == 0 {
		return ""
	}
	md = md[len("<h1>") : len(md)-len("</h1>")]
	// Strip out bad HTML
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("class", "id").Globally()
	outHTML := string(policy.SanitizeBytes(md))
	// Post-processing
	outHTML = paragraphReg.ReplaceAllString(outHTML, "$1")
	outHTML = strings.TrimRightFunc(outHTML, unicode.IsSpace)

	return outHTML
}

func postTitle(content, friendlyId string) string {
	const maxTitleLen = 80

	content = stripHTMLWithoutEscaping(content)

	content = strings.TrimLeftFunc(stripmd.Strip(content), unicode.IsSpace)
	eol := strings.IndexRune(content, '\n')
	blankLine := strings.Index(content, "\n\n")
	if blankLine != -1 && blankLine <= eol && blankLine <= assumedTitleLen {
		return strings.TrimSpace(content[:blankLine])
	} else if utf8.RuneCountInString(content) <= maxTitleLen {
		return content
	}
	return friendlyId
}

// TODO: fix duplicated code from postTitle. postTitle is a widely used func we
// don't have time to investigate right now.
func friendlyPostTitle(content, friendlyId string) string {
	const maxTitleLen = 80

	content = stripHTMLWithoutEscaping(content)

	content = strings.TrimLeftFunc(stripmd.Strip(content), unicode.IsSpace)
	eol := strings.IndexRune(content, '\n')
	blankLine := strings.Index(content, "\n\n")
	if blankLine != -1 && blankLine <= eol && blankLine <= assumedTitleLen {
		return strings.TrimSpace(content[:blankLine])
	} else if eol == -1 && utf8.RuneCountInString(content) <= maxTitleLen {
		return content
	}
	title, truncd := parse.TruncToWord(parse.PostLede(content, true), maxTitleLen)
	if truncd {
		title += "..."
	}
	return title
}

// Strip HTML tags with bluemonday's StrictPolicy, then unescape the HTML
// entities added in by sanitizing the content.
func stripHTMLWithoutEscaping(content string) string {
	return html.UnescapeString(bluemonday.StrictPolicy().Sanitize(content))
}

func getSanitizationPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("src", "style").OnElements("iframe", "video", "audio")
	policy.AllowAttrs("src", "type").OnElements("source")
	policy.AllowAttrs("frameborder", "width", "height").Matching(bluemonday.Integer).OnElements("iframe")
	policy.AllowAttrs("allowfullscreen").OnElements("iframe")
	policy.AllowAttrs("controls", "loop", "muted", "autoplay").OnElements("video")
	policy.AllowAttrs("controls", "loop", "muted", "autoplay", "preload").OnElements("audio")
	policy.AllowAttrs("target").OnElements("a")
	policy.AllowAttrs("title").OnElements("abbr")
	policy.AllowAttrs("style", "class", "id").Globally()
	policy.AllowAttrs("alt").OnElements("img")
	policy.AllowElements("header", "footer")
	policy.AllowAttrs("method", "action").OnElements("form")
	policy.AllowAttrs("type", "name", "value", "placeholder").OnElements("input")
	policy.AllowURLSchemes("http", "https", "mailto", "xmpp", "gopher", "gophers", "gemini", "spartan")
	return policy
}

func sanitizePost(content string) string {
	return strings.Replace(content, "<", "&lt;", -1)
}

// postDescription generates a description based on the given post content,
// title, and post ID. This doesn't consider a V2 post field, `title` when
// choosing what to generate. In case a post has a title, this function will
// fail, and logic should instead be implemented to skip this when there's no
// title, like so:
//
//	var desc string
//	if title == "" {
//	    desc = postDescription(content, title, friendlyId)
//	} else {
//	    desc = shortPostDescription(content)
//	}
func postDescription(content, title, friendlyId string) string {
	maxLen := 140

	if content == "" {
		content = "WriteFreely is a painless, simple, federated blogging platform."
	} else {
		fmtStr := "%s"
		truncation := 0
		if utf8.RuneCountInString(content) > maxLen {
			// Post is longer than the max description, so let's show a better description
			fmtStr = "%s..."
			truncation = 3
		}

		if title == friendlyId {
			// No specific title was found; simply truncate the post, starting at the beginning
			content = fmt.Sprintf(fmtStr, strings.Replace(stringmanip.Substring(content, 0, maxLen-truncation), "\n", " ", -1))
		} else {
			// There was a title, so return a real description
			blankLine := strings.Index(content, "\n\n")
			if blankLine < 0 {
				blankLine = 0
			}
			truncd := stringmanip.Substring(content, blankLine, blankLine+maxLen-truncation)
			contentNoNL := strings.Replace(truncd, "\n", " ", -1)
			content = strings.TrimSpace(fmt.Sprintf(fmtStr, contentNoNL))
		}
	}

	return content
}

func shortPostDescription(content string) string {
	maxLen := 140
	fmtStr := "%s"
	truncation := 0
	if utf8.RuneCountInString(content) > maxLen {
		// Post is longer than the max description, so let's show a better description
		fmtStr = "%s..."
		truncation = 3
	}
	return strings.TrimSpace(fmt.Sprintf(fmtStr, strings.Replace(stringmanip.Substring(content, 0, maxLen-truncation), "\n", " ", -1)))
}

func handleRenderMarkdown(app *App, w http.ResponseWriter, r *http.Request) error {
	if !IsJSON(r) {
		return impart.HTTPError{Status: http.StatusUnsupportedMediaType, Message: "Markdown API only supports JSON requests"}
	}

	in := struct {
		CollectionURL string `json:"collection_url"`
		RawBody       string `json:"raw_body"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&in)
	if err != nil {
		log.Error("Couldn't parse markdown JSON request: %v", err)
		return ErrBadJSON
	}

	body := in.RawBody
	if in.CollectionURL != "" {
		body = strings.Replace(body, shortCodeMore, `<a href="/">Read more...</a>`, 1)
	}
	rendered := applyMarkdown([]byte(in.RawBody), in.CollectionURL, app.cfg)
	out := struct {
		Body string `json:"body"`
	}{
		Body: rendered,
	}

	return impart.WriteSuccess(w, out, http.StatusOK)
}
