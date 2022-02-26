// Provides a lexer for `*.mod` files in CK3
package ck3_parser

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

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const eof = -1 // end of file constant

////////////////////////////////////////////////////////////////
// TOKENS //////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// tokenType is the index of the constant enum for T_* constants
type tokenType int

// String return the token type's description or an error if it doesn't exist
func (t tokenType) String() string {
	switch t {
	case tArrayEnd:
		return "end of object"
	case tArrayStart:
		return "start of object"
	case tComment:
		return "comment"
	case tDefinition:
		return "definition"
	case tEof:
		return "end of file"
	case tError:
		return "error"
	case tKey:
		return "identifier key"
	case tNumber:
		return "number"
	case tPath:
		return "folder path"
	case tString:
		return "string"
	case tValueEnd:
		return "end of value"
	case tValueStart:
		return "start of value"
	case tVersion:
		return "version number"
	default:
		panic(fmt.Sprintf("Descriptor lexer error: unknown token type"))
		// TODO: Throw error if we can have to invalid token types? Maybe we don't want to crash the program on bad lexing
		//err = errors.New("Lexer error: unknown token type")
	}
}

// t* constants are the types of token returned by the descriptor's lexer
const (
	tArrayEnd   tokenType = iota // }
	tArrayStart                  // {
	tComment                     // # a comment
	tDefinition                  // = preceeds a value
	tEof                         // end of file
	tError                       // a stored lex error
	tKey                         // an identifier: name, version, tags, etc. Followed by =
	tPath                        // a system folder path: "mod/mymod", "C:/users/me/Paradox/CK3/mod/mymod", etc.
	tNumber                      // a number
	tString                      // a text string
	tValueEnd                    // "
	tValueStart                  // "
	tVersion                     // an version number: uses semantic versioning (MAJOR.MINOR.PATCH). Wildcards (*) may be used to define a range of versions.
)

// toke describes a token with its type and its value
type token struct {
	typ tokenType
	val string
}

// this is the state function that is called recursively. It simply returns the next lexing function, or `nil for EOF`
type stateFn func(*lexer) stateFn

// `lex` is the main function. It creates a new lexer, runs it and returns the `out` send-only channel for the parser to process
func Lex(inputReader io.Reader) chan token {
	l := lexer{
		in:     bufio.NewReader(inputReader),
		out:    make(chan token),
		backup: make([]rune, 0, 4),
	}
	go l.run()
	return l.out
}

////////////////////////////////////////////////////////////////
// LEXER ///////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// lexer holds the state of the scanner.
type lexer struct {
	in     io.RuneReader
	out    chan token // channel emitting scanned tokens
	buf    []rune     // running buffer for current lexeme
	backup []rune     // backup buffer for current lexeme
	err    error      //used to store error messages used to send tError tokens on the out channel
}

// `accept` checks if the next rune is included in the string parameter
// If yes, it keeps the rune by adding it to the read buffer and returns true
// If not, it un-reads the rune and returns false
func (l *lexer) accept(validRunes string) bool {
	rune := l.next()
	if strings.ContainsRune(validRunes, rune) {
		l.keep(rune)
		return true
	} else {
		l.unread(rune)
		return false
	}
}

// `acceptRun` accepts runes as long as they are included in the string parameter
func (l *lexer) acceptRun(validRunes string) bool {
	none := true
	for l.accept(validRunes) {
		none = false
	}
	return !none
}

// `emit` passes a token back to the client on the `out` channel
func (l *lexer) emit(t tokenType) {
	l.out <- token{t, string(l.buf)} // send the token in the out channel
	l.buf = l.buf[0:0]               // clear the buffer
}

// `keep` writes the rune in the current read buffer
func (l *lexer) keep(r rune) {
	if l.buf == nil {
		l.buf = make([]rune, 0, 18)
	}
	l.buf = append(l.buf, r)
}

// next reads the next rune and returns it
func (l *lexer) next() rune {
	// if we have unread some runes, we set the run to be read as the last (oldest) one in the backup buffer
	// then we remove the rune  we just read from the backup buffer
	if len(l.backup) > 0 {
		r := l.backup[len(l.backup)-1]
		l.backup = l.backup[:len(l.backup)-1]
		return r
	}
	r, _, err := l.in.ReadRune()
	switch err {
	case io.EOF:
		return eof
	case nil:
		return r
	default:
		l.err = err
		return eof
	}
}

// peek() returns the next rune without affecting the read buffer
func (l *lexer) peek() rune {
	r := l.next()
	l.unread(r)
	return r
}

// `run` runs the lexer as a recursive function that acts as a state machine
// it emits tokens in an `out` channel and keeps lexing until it reaches EOF which returns a `nil` state
func (l *lexer) run() {
	defer close(l.out) // close the out channel when exiting the lexer
	for state := lexRoot; state != nil; {
		state = state(l)
		if l.err != nil {
			state = lexErrorf("read error: %s", l.err)
		}
	}
}

// unread adds the current rune to the backup buffer
func (l *lexer) unread(r rune) {
	l.backup = append(l.backup, r)
}

////////////////////////////////////////////////////////////////
// STATES //////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////

// `lexArray` lexes the first value and keep lexing values until we meet the end curly bracket '}'
// func lexArray() stateFn {
// }

// `lexComment` lexes a comment until it meets a new line '\n'
func lexComment(l *lexer) stateFn {
	switch r := l.next(); r {
	case '\n':
		l.emit(tComment)
		return lexRoot
	case eof:
		l.emit(tComment)
		return nil
	default:
		l.keep(r)
		return lexComment
	}
}

// `lexError` returns the token as a tError tokenType with the error string as value
func lexErrorf(errMsg string, args ...interface{}) stateFn {
	return func(l *lexer) stateFn {
		l.out <- token{tError, fmt.Sprintf(errMsg, args...)}
		return nil
	}
}

func lexKey(l *lexer) stateFn {
	switch r := l.next(); r {
	case ' ', '\t', '=':
		l.emit(tKey)
		// DEBUG: this should return to lexRoot
		return nil
		// return lexRoot
	case eof:
		l.emit(tKey)
		return nil
	default:
		l.keep(r)
		return lexKey
	}
}

func lexRoot(l *lexer) stateFn {
	r := l.next()
	switch {
	case r == eof:
		return nil
	case r == '#':
		return lexComment
	case r == '=': // end the lexer for now TODO: remove this and handle value
		return nil
	case unicode.IsLetter(r):
		l.keep(r)
		return lexKey
	// case r == '=':
	// 	return lexRightHandSide(r)
	// case unicode.IsSpace(r):
	// return skipEmpty
	// case unicode.IsPrint(r):
	// return lexKey(r)
	default:
		return lexErrorf("unexpected rune in lexRoot: %c", r)
	}
}

// `lexRightHandSide` lexes the right hand side of a definition. It can either be a value or an array of values
// func lexRightHandSide() stateFn {
// 	// skip whitespaces and read next rune
// 	// if '"" then lex the value
// 	// if '{" then lex the value array

// 	return func(l *lexer) stateFn {
// 		switch r := l.next(); r {
// 		case '"':
// 			l.next()
// 			return lexValue()
// 		case '{':
// 			l.next()
// 			// return lexArray()
// 		default:
// 			return lexErrorf("unexpected rune in lexRightHandSide: %c", r)
// 		}
// 	}
// }

// `lexValue` lexes an entire value until we meet the end double quote '""
// func lexValue() stateFn {
// }
