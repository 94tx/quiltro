package gemtext

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type NodeType uint16
type ParserState uint16

const (
	PSNone         ParserState = iota
	PSMetadata                 = iota
	PSPreformatted             = iota
	PSBlockquote               = iota
	PSList                     = iota
)

const (
	NodeNil               NodeType = iota
	NodePara                       = iota
	NodeH1                         = iota
	NodeH2                         = iota
	NodeH3                         = iota
	NodeLink                       = iota
	NodeListStart                  = iota
	NodeListText                   = iota
	NodeListEnd                    = iota
	NodePreformattedStart          = iota
	NodePreformattedAlt            = iota
	NodePreformattedText           = iota
	NodePreformattedEnd            = iota
	NodeBlockquoteStart            = iota
	NodeBlockquoteText             = iota
	NodeBlockquoteEnd              = iota
)

type Node struct {
	Type NodeType
	Text string
	URL  string
}

type Parser struct {
	reader     io.Reader
	scanner    *bufio.Scanner
	state      ParserState
	leftover   string
	nodeExists bool
	node       Node
}

func NewParser(r io.Reader) Parser {
	st := PSNone
	return Parser{
		r,
		bufio.NewScanner(r),
		st,
		"",
		false,
		Node{},
	}
}

func NewParserFromString(s string) Parser {
	return NewParser(strings.NewReader(s))
}
func NewParserFromBytes(s []byte) Parser {
	return NewParser(bytes.NewReader(s))
}

func (gp *Parser) Node() Node {
	if gp.nodeExists {
		return gp.node
	} else {
		return Node{NodeNil, "", ""}
	}
}

func (gp *Parser) Next() bool {
	var line string

	if gp.leftover != "" {
		line = gp.leftover
		gp.leftover = ""
	} else {
		if cont := gp.scanner.Scan(); !cont {
			return false
		}
		line = gp.scanner.Text()
		if line == "" {
			if gp.state != PSNone {
				gp.nodeExists = true
				switch gp.state {
				case PSPreformatted:
					gp.node = Node{NodePreformattedEnd, "", ""}
				case PSBlockquote:
					gp.node = Node{NodeBlockquoteEnd, "", ""}
				case PSList:
					gp.node = Node{NodeListEnd, "", ""}
				}
				gp.state = PSNone
			} else {
				gp.nodeExists = false
			}

			return true
		}
	}

	gp.nodeExists = true
	gp.node = Node{}

	switch gp.state {
	case PSNone:
		if strings.HasPrefix(line, "###") {
			gp.node = Node{NodeH3, strings.TrimSpace(line[3:]), ""}
		} else if strings.HasPrefix(line, "##") {
			gp.node = Node{NodeH2, strings.TrimSpace(line[2:]), ""}
		} else if strings.HasPrefix(line, "#") {
			gp.node = Node{NodeH1, strings.TrimSpace(line[1:]), ""}
		} else if strings.HasPrefix(line, "*") {
			gp.state = PSList
			gp.node = Node{NodeListStart, strings.TrimSpace(line[1:]), ""}
		} else if strings.HasPrefix(line, "=>") {
			text := strings.TrimSpace(line[2:])
			link, text, found := strings.Cut(text, " ")
			if found {
				gp.node = Node{NodeLink, text, link}
			} else {
				gp.node = Node{NodeLink, link, link}
			}
		} else if strings.HasPrefix(line, "```") {
			gp.state = PSPreformatted
			_, alt, _ := strings.Cut(line, "```")
			if alt != "" {
				gp.node = Node{NodePreformattedAlt, strings.TrimSpace(alt), ""}
			} else {
				gp.node = Node{NodePreformattedStart, "", ""}
			}
		} else if strings.HasPrefix(line, ">") {
			gp.state = PSBlockquote
			gp.node = Node{NodeBlockquoteStart, strings.TrimSpace(line[1:]), ""}
		} else {
			gp.node = Node{NodePara, line, ""}
		}
	case PSList:
		if strings.HasPrefix(line, "*") {
			gp.node = Node{NodeListText, strings.TrimSpace(line[1:]), ""}
		} else {
			gp.state = PSNone
			gp.leftover = line
			gp.node = Node{NodeListEnd, "", ""}
		}
	case PSPreformatted:
		if !strings.HasPrefix(line, "```") {
			gp.node = Node{NodePreformattedText, line, ""}
		} else {
			gp.state = PSNone
			gp.node = Node{NodePreformattedEnd, "", ""}
		}
	case PSBlockquote:
		if strings.HasPrefix(line, ">") {
			gp.node = Node{NodeBlockquoteText, strings.TrimSpace(line[1:]), ""}
		} else {
			gp.state = PSNone
			gp.leftover = line
			gp.node = Node{NodeBlockquoteEnd, "", ""}
		}
	}

	return true
}

func (gp *Parser) WriteDocument(w io.Writer, fn func(n Node, w io.Writer) error) error {
	for gp.Next() {
		if node := gp.Node(); node.Type != NodeNil {
			err := fn(node, w)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
