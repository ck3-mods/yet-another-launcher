package ck3_parser

import (
	"bytes"
	"fmt"
	"io"
)

const indent = "  " // indentation for the prettifier

type nodeType int

const (
	nArray nodeType = iota
	nAssignment
	nComment
	nError
	nRoot
	nValue
)

type node interface {
	Type() nodeType
	parse(*parser) error
	pretty(io.Writer, string) error
}

////////////////////////////////////////////////////////////////
// ROOT NODE ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////
type rootNode struct {
	children []node
}

func newRootNode() node {
	return &rootNode{children: make([]node, 0, 8)}
}

func (n *rootNode) Type() nodeType {
	return nRoot
}

func (n *rootNode) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	for _, child := range n.children {
		fmt.Fprintf(&buf, "%s, ", child)
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 2)
	}
	buf.WriteString("}")
	return buf.String()
}

func (n *rootNode) addChild(child node) {
	if n.children == nil {
		n.children = make([]node, 0, 8)
	}
	n.children = append(n.children, child)
}

func (n *rootNode) parse(p *parser) error {
	for {
		tok := p.next()
		switch tok.typ {
		case tError:
			return fmt.Errorf("parse error: saw lex error while parsing root node: %v", tok)
		case tEof:
			return nil
		case tComment:
			n.addChild(&commentNode{tok.val})
			// discard comments
		case tKey:
			keyNode := &assignmentNode{name: tok.val}
			if err := keyNode.parse(p); err != nil {
				return err
			}
			n.addChild(keyNode)
		default:
			return fmt.Errorf("parse error: unexpected token type %v while parsing root node", tok.typ)
		}
	}
}

func (n *rootNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sroot:\n", prefix)
	for _, child := range n.children {
		if err := child.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////
// ARRAY NODE //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// `arrayNode` represents an array of values
type arrayNode []node

func (n *arrayNode) Type() nodeType {
	return nArray
}

func (n *arrayNode) parse(p *parser) error {
	if p.peek().typ == tArrayEnd {
		p.next()
		return nil
	}

	if valueNode, err := p.parseValue(); err != nil {
		return err
	} else {
		*n = append(*n, valueNode)
	}

	switch tok := p.peek(); tok.typ {
	case tArrayEnd:
		p.next()
		return nil
	default:
		return n.parse(p)
	}
}

func (l *arrayNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sarray:\n", prefix)
	for _, n := range *l {
		if err := n.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////
// ASSIGNMENT NODE /////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// `assignementNode` represents an assignement of a value, or array of value, to an identifier key
// the name of the identifier key is in the `name` field, and the value in the `value` field
type assignmentNode struct {
	name  string
	value node
}

func (n *assignmentNode) Type() nodeType {
	return nAssignment
}

func (n *assignmentNode) String() string {
	return fmt.Sprintf("{assign: name=%s, val=%v}", n.name, n.value)
}

func (n *assignmentNode) parse(p *parser) error {
	tok := p.next()
	switch tok.typ {
	case tError:
		return fmt.Errorf("parse error: saw lex error while parsing assignment node: %v", tok.val)
	case tEof:
		return fmt.Errorf("parse error: unexpected eof in assignment node")
	case tDefinition: // exit if we find the right token
		break
	default:
		return fmt.Errorf("parse error: unexpected %v token after identifier key, expected =", tok.typ)
	}

	value, err := p.parseValue()
	if err != nil {
		return err
	}
	n.value = value
	return nil
}

func (n *assignmentNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sassign:\n", prefix)
	fmt.Fprintf(w, "%s%sname:\n", prefix, indent)
	fmt.Fprintf(w, "%s%s%s%s\n", prefix, indent, indent, n.name)
	fmt.Fprintf(w, "%s%svalue:\n", prefix, indent)
	if err := n.value.pretty(w, prefix+indent+indent); err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////
// COMMENT NODE //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// `commentNode` represents a string value, held in the body field
type commentNode struct {
	body string
}

func (n *commentNode) Type() nodeType {
	return nComment
}

func (n *commentNode) parse(p *parser) error {
	return nil
}

func (n *commentNode) String() string {
	return fmt.Sprintf("{comment: %s}", n.body)
}

func (n *commentNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%scomment:\n", prefix)
	fmt.Fprintf(w, "%s%s%s\n", prefix, indent, n.body) // comments are one-liner in descriptor files
	return nil
}

////////////////////////////////////////////////////////////////
// VALUE NODE //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// `valueNode` represents a string value, held in the name field
type valueNode struct {
	name string
}

func (n *valueNode) Type() nodeType {
	return nValue
}

func (n *valueNode) parse(p *parser) error {
	tok := p.next()
	if tok.typ != tValue {
		return fmt.Errorf("unexpected %s token when parsing variable", tok.typ)
	}
	n.name = tok.val
	return nil
}

func (n *valueNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%svalue:\n", prefix)
	fmt.Fprintf(w, "%s%s\n", prefix+indent, n.name)
	return nil
}
