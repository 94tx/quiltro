package gemtext

import (
	"fmt"
	"html"
	"io"
	"mime"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type HTMLWriter struct {
	Level int
	
	hids map[string]int
}

var spaceRegex = regexp.MustCompile(`(\pP|\pC|\pZ)+`)

func (h *HTMLWriter) WriteNode(n Node, w io.Writer) error {
	if h.Level == 0 {
		h.Level = 1
	}
	if h.hids == nil {
		h.hids = make(map[string]int)
	}
	
	switch n.Type {
	case NodePara:
		_, err := fmt.Fprintf(w, "<p>%s</p>\n", html.EscapeString(n.Text))
		if err != nil {
			return err
		}
	case NodeH1, NodeH2, NodeH3:
		id := spaceRegex.ReplaceAllString(strings.ToLower(n.Text), "-")
		h.hids[id] += 1

		if h.hids[id] > 1 {
			id = fmt.Sprintf("%s-%d", id, h.hids[id] - 1)
		}
		
		level := int(n.Type - NodeH1) + h.Level
		_, err := fmt.Fprintf(
			w,
			`<h%d id="%s"><a href="#%s">%s</a></h%d>\n`,
			level,
			id,
			id,
			html.EscapeString(n.Text),
			level,
		)
		if err != nil {
			return err
		}
	case NodeLink:
		url, err := url.Parse(n.URL)
		if err != nil {
			return err
		}

		if filepath.Ext(url.Path) == ".gmi" && url.Scheme == "" {
			// convert local gemini links to html, since they're going to be
			// served by the same application
			url.Path = strings.TrimSuffix(url.Path, ".gmi") + ".html"
		} else if url.Scheme == "gemini" {
			url.Scheme = ""
			curl := url.String()
			url, err = url.Parse("http://portal.mozz.us/gemini" + curl[1:])
		}
		if err != nil {
			return err
		}

		mime := mime.TypeByExtension(path.Ext(url.Path))
		if strings.HasPrefix(mime, "image/") {
			_, err := fmt.Fprintf(
				w,
				"<figure>\n<a href=\"%s\"><img src=\"%s\" /></a>\n",
				url, url,
			)
			if err != nil {
				return err
			}
			if n.URL != n.Text {
				_, err := fmt.Fprintf(
					w,
					"<figcaption>%s</figcaption>\n",
					html.EscapeString(n.Text),
				)
				if err != nil {
					return err
				}
			}
			_, err = fmt.Fprintln(w, "</figure>")
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintf(w, "<p class=\"link\"><a href=\"%s\"", url)
			if err != nil {
				return err
			}
			if url.Scheme != "" {
				_, err := fmt.Fprint(w, " class=\"ext-link\"")
				if err != nil {
					return err
				}
			}
			if n.URL != n.Text {
				_, err = fmt.Fprintf(
					w,
					" title=\"%s\">%s (%s)</a></p>\n",
					html.EscapeString(n.Text),
					html.EscapeString(n.Text),
					html.EscapeString(n.URL),
				)
			} else {
				_, err = fmt.Fprintf(
					w,
					" title=\"%s\">%s</a></p>\n",
					html.EscapeString(n.Text),
					html.EscapeString(n.Text),
				)
			}
			if err != nil {
				return err
			}
		}
	case NodeListStart:
		_, err := fmt.Fprintf(w, "<ul>\n<li>%s</li>\n",
			html.EscapeString(n.Text))
		if err != nil {
			return err
		}
	case NodeListText:
		_, err := fmt.Fprintf(w, "<li>%s</li>\n", html.EscapeString(n.Text))
		if err != nil {
			return err
		}
	case NodeListEnd:
		_, err := fmt.Fprintln(w, "</ul>")
		if err != nil {
			return err
		}
	case NodePreformattedStart:
		_, err := fmt.Fprint(w, "<pre>")
		if err != nil {
			return err
		}
	case NodePreformattedAlt:
		alt := html.EscapeString(n.Text)
		_, err := fmt.Fprintf(
			w,
			"<pre aria-label=\"%s\">",
			alt,
		)
		if err != nil {
			return err
		}
	case NodePreformattedText:
		_, err := fmt.Fprintln(w, n.Text)
		if err != nil {
			return err
		}
	case NodePreformattedEnd:
		_, err := fmt.Fprintln(w, "</pre>")
		if err != nil {
			return err
		}
	case NodeBlockquoteStart:
		_, err := fmt.Fprintf(
			w,
			"<blockquote>\n<p>%s</p>\n",
			html.EscapeString(n.Text),
		)
		if err != nil {
			return err
		}
	case NodeBlockquoteText:
		_, err := fmt.Fprintf(w, "<p>%s</p>\n", html.EscapeString(n.Text))
		if err != nil {
			return err
		}
	case NodeBlockquoteEnd:
		_, err := fmt.Fprintln(w, "</blockquote>")
		if err != nil {
			return err
		}
	}

	return nil
}
