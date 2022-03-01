package ck3_parser

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
	eval() (interface{}, error)
	parse(*parser) error
	pretty(io.Writer, string) error
}

////////////////////////////////////////////////////////////////
// ROOT NODE ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////
type rootNode struct {
	children []node
}

// allocate a new root node
func newRootNode() node {
	return &rootNode{children: make([]node, 0, 8)}
}

// return the `nodeType`
func (n *rootNode) Type() nodeType {
	return nRoot
}

// print the node as a string
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

// add a child node to the root node. Allocate memory if no children exist
func (n *rootNode) addChild(child node) {
	if n.children == nil {
		n.children = make([]node, 0, 8)
	}
	n.children = append(n.children, child)
}

// parse the node while there are tokens in the parser
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

// pretty print the node to the io.Writer passed as parameter, with an optional prefix
func (n *rootNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sroot:\n", prefix)
	for _, child := range n.children {
		if err := child.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	return nil
}

// evaluate the node and its children and return an object representing the parse tree
// the format is { key1: value1, key2: [value2, value3] }
func (n *rootNode) eval() (interface{}, error) {
	values := make(map[string]interface{}, 0)

	for _, child := range n.children {
		if child.Type() != nAssignment { // we only care about assignements, not comments
			continue
		}
		childValue, _ := child.eval()                            // get the map{key: value} child
		childValueMap, ok := childValue.(map[string]interface{}) // cast the value to get the map{key: value} child
		if !ok {
			log.Printf("Error in the value type: %v", child.Type())
		}
		for k, v := range childValueMap {
			values[k] = v // we add it to the parsed value map
		}
	}
	return values, nil
}

////////////////////////////////////////////////////////////////
// ARRAY NODE //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// an array of value nodes
type arrayNode []node

// return the `nodeType`
func (n *arrayNode) Type() nodeType {
	return nArray
}

// print the node as a string
func (n *arrayNode) String() string {
	return fmt.Sprint("{array }")
}

// parse the array node while there are value to be parsed. Stop on `tArrayEnd` token
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

// pretty print the node to the io.Writer passed as parameter, with an optional prefix
func (l *arrayNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sarray:\n", prefix)
	for _, n := range *l {
		if err := n.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	return nil
}

// evaluate the array node, returning an array of string values
func (n *arrayNode) eval() (interface{}, error) {
	array := make([]string, len(*n))
	for i, valueNode := range *n {
		value, _ := valueNode.eval()
		stringValue, isString := value.(string)
		if !isString {
			fmt.Printf("node's value is not a string but %T", value)
		}
		array[i] = string(stringValue)
	}
	return array, nil
}

////////////////////////////////////////////////////////////////
// ASSIGNMENT NODE /////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// an assignement of a value, or array of value, to an identifier key
// the name of the identifier key is in the `name` field, and the value in the `value` field
type assignmentNode struct {
	name  string
	value node
}

// return the `nodeType`
func (n *assignmentNode) Type() nodeType {
	return nAssignment
}

// print the node as a string
func (n *assignmentNode) String() string {
	return fmt.Sprintf("{assign: name=%s, val=%v}", n.name, n.value)
}

// parse the assignment node, its key and contained value.s
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

// pretty print the node to the io.Writer passed as parameter, with an optional prefix
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

// evaluate the assignment node, returning an object with the node's name as key and the child value.s as value
func (n *assignmentNode) eval() (interface{}, error) {
	assignement := make(map[string]interface{})
	key := n.name
	value, _ := n.value.eval()
	assignement[key] = value
	return assignement, nil
}

////////////////////////////////////////////////////////////////
// COMMENT NODE //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// a comment, as a string value held in the body field
type commentNode struct {
	body string
}

// return the `nodeType`
func (n *commentNode) Type() nodeType {
	return nComment
}

func (n *commentNode) parse(p *parser) error {
	return nil
}

// print the node as a string
func (n *commentNode) String() string {
	return fmt.Sprintf("{comment: %s}", n.body)
}

// pretty print the node to the io.Writer passed as parameter, with an optional prefix
func (n *commentNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%scomment:\n", prefix)
	fmt.Fprintf(w, "%s%s%s\n", prefix, indent, n.body) // comments are one-liner in descriptor files
	return nil
}

func (n *commentNode) eval() (interface{}, error) {
	return nil, nil
}

////////////////////////////////////////////////////////////////
// VALUE NODE //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// a string value, held in the name field
type valueNode struct {
	name string
}

// return the `nodeType`
func (n *valueNode) Type() nodeType {
	return nValue
}

// print the node as a string
func (n *valueNode) String() string {
	return fmt.Sprintf("{value: %s}", n.name)
}

// parse the node and return its value as the node's name field
func (n *valueNode) parse(p *parser) error {
	tok := p.next()
	if tok.typ != tValue {
		return fmt.Errorf("unexpected %s token when parsing variable", tok.typ)
	}
	n.name = tok.val
	return nil
}

// pretty print the node to the io.Writer passed as parameter, with an optional prefix
func (n *valueNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%svalue:\n", prefix)
	fmt.Fprintf(w, "%s%s\n", prefix+indent, n.name)
	return nil
}

func (n *valueNode) eval() (interface{}, error) {
	return n.name, nil
}
