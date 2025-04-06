package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
)

type TokenType uint8

const (
	group   TokenType = iota
	bracket TokenType = iota
	or      TokenType = iota
	repeat  TokenType = iota
	literal TokenType = iota
	bundle  TokenType = iota
)

type Token struct{
	tokenType TokenType
	value interface{}
}

type Parser struct {
	i int // current index in the regex string
	tokens []Token
}

type State struct {
	start bool
	end bool
	transitions map[uint8][]*State
}

func exitWithMsg(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func parseGroup(rgx string, parser *Parser) {
	parser.i++
	for parser.i != len(rgx) && rgx[parser.i] != ')' {
		process(rgx, parser)
		parser.i++
	}
	if parser.i == len(rgx) {
		exitWithMsg("Error: there is no ')' to end group")
	}
}

func parseBracket(rgx string, parser *Parser) {
	parser.i++
	var literals []string
	for parser.i != len(rgx) && rgx[parser.i] != ']' {
		c := rgx[parser.i]
		if c == '-' {
			if parser.i + 1 == len(rgx) || rgx[parser.i + 1] == ']' {
				exitWithMsg("Error: a range must has an end")
			}
			if len(literals) == 0 || len(literals[len(literals) - 1]) != 1 {
				exitWithMsg("Error: a range must has a start")
			}
			nextChar := rgx[parser.i + 1]
			prevChar := literals[len(literals) - 1][0]
			literals[len(literals) - 1] = fmt.Sprintf("%c%c", prevChar, nextChar)
			parser.i++
		} else {
			literals = append(literals, fmt.Sprintf("%c", c))
		}
		parser.i++
	}
	if parser.i == len(rgx) {
		exitWithMsg("Error: there is no ']' to end bracket")
	}
	literalsSet := make(map[uint8]bool)
	for _, lit := range literals {
		if lit[0] > lit[len(lit) - 1] {
			exitWithMsg("Error: range start must be less that range end")
		}
		for c := lit[0]; c <= lit[len(lit) - 1]; c++ {
			literalsSet[c] = true
		}
	}

	parser.tokens = append(parser.tokens, Token{ 
		tokenType: bracket,
		value:     literalsSet,
	})
}

func parseOr(rgx string, parser *Parser) {
	rhsParser := &Parser{
		i: parser.i,
		tokens: []Token{},
	}
	rhsParser.i += 1
	for rhsParser.i < len(rgx) && rgx[rhsParser.i] != ')' {
		process(rgx, rhsParser)
		rhsParser.i += 1
	}
	left := Token{
		tokenType: bundle,
		value: parser.tokens, 
	}
	right := Token{ 
		tokenType: bundle,
		value: rhsParser.tokens,
	}
	parser.tokens = []Token{{ 
		tokenType: or,
		value: []Token{left, right},
	}}
	parser.i = rhsParser.i 
}

const INF = -1

func parseRepeat(rgx string, parser *Parser) {
	c := rgx[parser.i]
	var mx, mn int
	if c == '*' {
		mx = INF
		mn = 0
	} else if c == '+' {
		mx = INF
		mn = 1 
	} else {
		mx = 1
		mn = 0
	}
	if len(parser.tokens) == 0 {
		exitWithMsg(fmt.Sprintf("Error: '%c' must has something before it", c))
	}
	lastToken := parser.tokens[len(parser.tokens) - 1]
	parser.tokens[len(parser.tokens) - 1] = Token {
		tokenType: repeat,
		value: struct {
			token Token
			min int
			max int
		}{ 
			token: lastToken,
			min:   mn,
			max:   mx,
		},
	}
}

func parseCustomRepeat(rgx string, parser *Parser) {
	l := parser.i + 1
	for parser.i < len(rgx) && rgx[parser.i] != '}' {
		parser.i++
	}
	r := parser.i

	if parser.i == len(rgx) {
		exitWithMsg("Error: there is no '}' to end the custom repeat")
	}

	leftRight := strings.Split(rgx[l:r], ",")
	if leftRight[0] == "" {
		exitWithMsg("Error: the custom repeat must has a min value")
	}

	var mx, mn int
	if len(leftRight) == 1 {
		val, err := strconv.Atoi(leftRight[0])
		if err != nil {
			exitWithMsg(err.Error())
		}
		mx = val
		mn = val
	} else if len(leftRight) == 2 {
		rightVal := INF
		leftVal, err := strconv.Atoi(leftRight[0])
		if err != nil {
			exitWithMsg(err.Error())
		}
		if leftRight[1] != "" {
			rightVal, err = strconv.Atoi(leftRight[1])
			if err != nil {
				exitWithMsg(err.Error())
			}
		}
		mn = leftVal
		mx = rightVal
	} else {
		exitWithMsg("Error: custom repeat must has two nums separated by a comma")
	}

	if len(parser.tokens) == 0 {
		exitWithMsg("Error: custom repeat must has something before it")
	}
	lastToken := parser.tokens[len(parser.tokens) - 1]
	parser.tokens[len(parser.tokens) - 1] = Token {
		tokenType: repeat,
		value: struct {
			token Token
			min int
			max int
		}{ 
			token: lastToken,
			min:   mn,
			max:   mx,
		},
	}
}

func process(rgx string, parser *Parser) {
	c := rgx[parser.i]
	if c == '(' {
		groupParser := &Parser{
			i: parser.i,
			tokens: []Token{},
		}
		parseGroup(rgx, groupParser)
		token := Token{
			tokenType: group,
			value: groupParser.tokens,
		}
		parser.tokens = append(parser.tokens, token)
		parser.i = groupParser.i
	} else if c == '[' {
		parseBracket(rgx, parser)
	} else if c == '|' {
		parseOr(rgx, parser)
	} else if c == '*' || c == '+' || c == '?' {
		parseRepeat(rgx, parser)
	} else if c == '{' {
		parseCustomRepeat(rgx, parser)
	} else {
		token := Token{
			tokenType: literal,
			value: c,
		}
		parser.tokens = append(parser.tokens, token)
	}
}

func parse(rgx string) []Token {
	parser := &Parser{
		i: 0,
		tokens: []Token{},
	}

	for parser.i < len(rgx) {
		process(rgx, parser)
		parser.i++
	}

	return parser.tokens
}

func tokenToNFA(token *Token) (*State, *State) {
	start := &State{
		transitions: map[uint8][]*State{},
	}
	end := &State{
		transitions: map[uint8][]*State{},
	}

	switch token.tokenType {
	case literal:
		start.transitions[token.value.(uint8)] = []*State{end}
	case or:
		token1 := token.value.([]Token)[0] 
		token2 := token.value.([]Token)[1] 
		start1, end1 := tokenToNFA(&token1)
		start2, end2 := tokenToNFA(&token2)
		start.transitions[epsilonValue] = []*State{start1, start2}
		end1.transitions[epsilonValue] = []*State{end}
		end2.transitions[epsilonValue] = []*State{end}
	case bracket:
		for ch := range token.value.(map[uint8]bool) {
			start.transitions[ch] = []*State{end}
		}
		
	default:
		exitWithMsg("Error: unknown token type")
	}

	return start, end
}

const epsilonValue uint8 = 0 // empty

func toNFA(tokens []Token) *State {
	startState, endState := tokenToNFA(&tokens[0])
	for i := 1; i < len(tokens); i++ {
		nextStart, nextEnd := tokenToNFA(&tokens[i])
		endState.transitions[epsilonValue] = append (
			endState.transitions[epsilonValue],
			nextStart,
		)
		endState = nextEnd
	}

	start := &State{
		start: true,
		transitions: map[uint8][]*State{
			epsilonValue: {startState},
		},
	}
	end := &State{
		end: true,
		transitions: map[uint8][]*State{},
	}

	endState.transitions[epsilonValue] = append (
		endState.transitions[epsilonValue],
		end,
	)

	return start
}

func match(state *State, input string, i int) bool {
	var c uint8
	if i == len(input) && state.end {
		return true
	} else if i != len(input) {
		c = input[i]
	} else {
		c = 1 
	}

	for _, nextState := range state.transitions[c] {
		if match(nextState, input, i + 1) {
			return true
		}
	}

	for _, nextEpsilonState := range state.transitions[epsilonValue] {
		if match(nextEpsilonState, input, i) {
			return true
		}
	}

	return false
}

func main() {
	tokens := parse(os.Args[1])
	startState := toNFA(tokens)
    fmt.Println(match(startState, os.Args[2], 0))
}
