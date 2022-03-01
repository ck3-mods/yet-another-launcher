package ck3_parser

import (
	"bytes"
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

/*
	Description of the *.mod format:
		identifier_key = "value"
		identifier_key = { "value1" "value2" ... }

	value format:
		string, path, number, or version number (X.X.*)

	identifier keys:
		|	key									|	required			|	type						|	definition
		------------------------------------------------------------------------------------------------
		|	name								|	yes						|	string	 				|	mod's name
		|	version	 						|	yes						|	version_number	|	mod's version
		|	tags								|	no						|	string array		|	array of steam category tags
		|	supported_version		|	yes*					|	version_number	|	game's supported version (required only for .mod inside mod folder not the descriptor)
		|	path								|	no?						|	path						|	mod's path (absolute or relative to user directory)
		|	remote_file_id			|	no						|	number	 	 			|	mod's remote file id (steam or paradox mods)
*/

// parser (little p) is an actual parser.  It actually does the parsing of a
// moon document.
type parser struct {
	root   node
	input  chan token
	backup []token
}

type ModDescriptor struct {
	name              string
	path              string
	remote_file_id    int
	supported_version string
	tags              []string
	version           string
}

// `valueParseFn` contain the valid node types for a value and returns a function to parse the corresponding node type
var valueParseFn = map[tokenType]func(p *parser) node{
	tArrayStart: func(p *parser) node { p.next(); return &arrayNode{} },
	tValue:      func(p *parser) node { return new(valueNode) },
}

// Reads a descriptor object from a given io.Reader. The io.Reader is advanced to EOF. The reader is not closed after reading,
// since it's an io.Reader and not an io.ReadCloser. In the event of error, the state that the source reader will be left in is undefined.
func Read(r io.Reader) (*ModDescriptor, error) {
	parseTree, err := parse(r)
	if err != nil {
		return nil, err
	}
	return &parseTree, nil
}

func ReadDescriptorFile(uri fyne.URI) (*ModDescriptor, error) {
	file, err := storage.Reader(uri)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Read(file)
}

func parse(reader io.Reader) (modDescriptor ModDescriptor, err error) {
	p := &parser{
		root:   newRootNode(),
		input:  lex(reader),
		backup: make([]token, 0, 8),
	}
	if err = p.run(); err != nil {
		return
	}

	////////////////////////////////////////////////////////////////
	// Pretty print the parse tree
	var buf bytes.Buffer
	p.root.pretty(&buf, "")
	if err != nil {
		return
	}
	fmt.Printf("Parse tree\n%s", buf.String())
	////////////////////////////////////////////////////////////////

	modDescriptor, err = convert(p.root)
	if err != nil {
		return
	}
	return
}

func convert(node) (modDescriptor ModDescriptor, err error) {
	// FIXME: NOT IMPLEMENTED
	return
}

func (p *parser) run() error {
	if p.root == nil {
		p.root = newRootNode()
	}
	return p.root.parse(p)
}

// `next` returns the next token and advances the input stream
func (p *parser) next() token {
	if len(p.backup) > 0 {
		oldestTok := p.backup[len(p.backup)-1]
		p.backup = p.backup[:len(p.backup)-1]
		return oldestTok
	}
SKIP_COMMENTS:
	tok, ok := <-p.input
	if !ok {
		return token{tEof, "eof"}
	}
	if tok.typ == tComment {
		goto SKIP_COMMENTS
	}
	return tok
}

// `parseValue` parses the next value.  To be executed in a context where we know we want something that is a value to come next, after an "=" sign
func (p *parser) parseValue() (node, error) {
	for {
		tok := p.peek()
		switch tok.typ {
		case tError:
			return nil, fmt.Errorf("parse error: saw lex error when looking for value: %v", tok.val)
		case tEof:
			return nil, fmt.Errorf("parse error: unexpected eof when looking for value")
		}

		parseValueFn, ok := valueParseFn[tok.typ]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %v token while looking for value", tok.typ)
		}
		n := parseValueFn(p)
		if err := n.parse(p); err != nil {
			return nil, err
		}
		return n, nil
	}
}

// `peek` returns the next token without affecting the current buffer
func (p *parser) peek() token {
	tok := p.next()
	p.unread(tok)
	return tok
}

// `unread` appends the token to the backup array, effectively causing `next` to re-read the token
func (p *parser) unread(t token) {
	if p.backup == nil {
		p.backup = make([]token, 0, 8)
	}
	p.backup = append(p.backup, t)
}
