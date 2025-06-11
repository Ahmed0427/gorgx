package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TokenType uint8

const (
	LITERAL  TokenType = iota
	CHAR_SET TokenType = iota
	REPEAT   TokenType = iota
	GROUP    TokenType = iota
	OR       TokenType = iota
)

const EPSILON uint8 = 0
const INFINITY int = 2147483647
const STRING_END = 1

type Token struct {
	tokenType TokenType
	value     interface{}
}

type Parser struct {
	i      int // current index in the regex string
	tokens []Token
}

type State struct {
	start       bool
	end         bool
	transitions []Transition
}
type Transition struct {
	symbol uint8
	states []*State
}

type RepeatData struct {
	token Token
	min   int
	max   int
}

func exit(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func parseGroup(rgx string, parser *Parser) {
	parser.i++
	for parser.i != len(rgx) && rgx[parser.i] != ')' {
		process(rgx, parser)
		parser.i++
	}
}

func parseCharSet(rgx string, parser *Parser) {
	parser.i++
	var literals []string
	for parser.i != len(rgx) && rgx[parser.i] != ']' {
		c := rgx[parser.i]
		if c == '-' {
			if (parser.i+1 != len(rgx) && rgx[parser.i+1] == ']') ||
				len(literals) == 0 {

				literals = append(literals, fmt.Sprintf("%c", c))
				parser.i++
				continue
			}
			if len(literals[len(literals)-1]) != 1 {
				exit("Error: a range must has a start")
			}
			nextChar := rgx[parser.i+1]
			prevChar := literals[len(literals)-1][0]
			literals[len(literals)-1] = fmt.Sprintf("%c%c", prevChar, nextChar)
			parser.i++
		} else {
			literals = append(literals, fmt.Sprintf("%c", c))
		}
		parser.i++
	}
	if parser.i == len(rgx) {
		exit("Error: there is no ']' to end bracket")
	}
	literalsSet := make(map[uint8]bool)
	for _, lit := range literals {
		if lit[0] > lit[len(lit)-1] {
			exit("Error: range start must be less that range end")
		}
		for c := lit[0]; c <= lit[len(lit)-1]; c++ {
			literalsSet[c] = true
		}
	}

	parser.tokens = append(parser.tokens, Token{
		tokenType: CHAR_SET,
		value:     literalsSet,
	})
}

func parseOr(rgx string, parser *Parser) {
	rhsParser := &Parser{
		i:      parser.i,
		tokens: []Token{},
	}
	rhsParser.i += 1
	for rhsParser.i < len(rgx) && rgx[rhsParser.i] != ')' {
		process(rgx, rhsParser)
		rhsParser.i += 1
	}
	left := Token{
		tokenType: GROUP,
		value:     parser.tokens,
	}
	right := Token{
		tokenType: GROUP,
		value:     rhsParser.tokens,
	}
	parser.tokens = []Token{{
		tokenType: OR,
		value:     []Token{left, right},
	}}
	parser.i = rhsParser.i
}

func parseRepeat(rgx string, parser *Parser) {
	c := rgx[parser.i]
	var mx, mn int
	if c == '*' {
		mx = INFINITY
		mn = 0
	} else if c == '+' {
		mx = INFINITY
		mn = 1
	} else {
		mx = 1
		mn = 0
	}
	if len(parser.tokens) == 0 {
		exit(fmt.Sprintf("Error: '%c' must has something before it", c))
	}
	lastToken := parser.tokens[len(parser.tokens)-1]
	parser.tokens[len(parser.tokens)-1] = Token{
		tokenType: REPEAT,
		value: RepeatData{
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
		exit("Error: there is no '}' to end the custom repeat")
	}

	leftRight := strings.Split(rgx[l:r], ",")
	if leftRight[0] == "" {
		exit("Error: the custom repeat must has a min value")
	}

	var mx, mn int
	if len(leftRight) == 1 {
		val, err := strconv.Atoi(leftRight[0])
		if err != nil {
			exit(err.Error())
		}
		mx = val
		mn = val
	} else if len(leftRight) == 2 {
		leftVal := 0
		rightVal := INFINITY
		var err error = nil
		if leftRight[0] != "" {
			leftVal, err = strconv.Atoi(leftRight[0])
			if err != nil {
				exit(err.Error())
			}
		}
		if leftRight[1] != "" {
			rightVal, err = strconv.Atoi(leftRight[1])
			if err != nil {
				exit(err.Error())
			}
		}
		if rightVal < leftVal {
			exit("Error: custom repeat must has max >= min")
		}
		if rightVal <= 0 {
			exit("Error: custom repeat must has max > 0")
		}
		mn = leftVal
		mx = rightVal
	} else {
		exit("Error: custom repeat must has two nums separated by a comma")
	}

	if len(parser.tokens) == 0 {
		exit("Error: custom repeat must has something before it")
	}
	lastToken := parser.tokens[len(parser.tokens)-1]
	parser.tokens[len(parser.tokens)-1] = Token{
		tokenType: REPEAT,
		value: RepeatData{
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
			i:      parser.i,
			tokens: []Token{},
		}
		parseGroup(rgx, groupParser)
		token := Token{
			tokenType: GROUP,
			value:     groupParser.tokens,
		}
		parser.tokens = append(parser.tokens, token)
		parser.i = groupParser.i
	} else if c == '[' {
		parseCharSet(rgx, parser)
	} else if c == '|' {
		parseOr(rgx, parser)
	} else if c == '*' || c == '+' || c == '?' {
		parseRepeat(rgx, parser)
	} else if c == '{' {
		parseCustomRepeat(rgx, parser)
	} else {
		token := Token{
			tokenType: LITERAL,
			value:     c,
		}
		parser.tokens = append(parser.tokens, token)
	}
}

func parse(rgx string) []Token {
	parser := &Parser{
		i:      0,
		tokens: []Token{},
	}

	for parser.i < len(rgx) {
		process(rgx, parser)
		parser.i++
	}

	return parser.tokens
}

func addTransition(a, b *State, symbol uint8) {
	for i := range a.transitions {
		if a.transitions[i].symbol != symbol {
			continue
		}
		a.transitions[i].states = append(a.transitions[i].states, b)
		return
	}

	a.transitions = append(a.transitions,
		Transition{symbol: symbol, states: []*State{b}},
	)
}

func tokenToNFA(token *Token) (*State, *State) {
	start := &State{
		transitions: []Transition{},
	}
	end := &State{
		transitions: []Transition{},
	}

	switch token.tokenType {
	case LITERAL:
		addTransition(start, end, token.value.(uint8))

	case OR:
		token1 := token.value.([]Token)[0]
		token2 := token.value.([]Token)[1]
		start1, end1 := tokenToNFA(&token1)
		start2, end2 := tokenToNFA(&token2)
		addTransition(start, start1, EPSILON)
		addTransition(start, start2, EPSILON)
		addTransition(end1, end, EPSILON)
		addTransition(end2, end, EPSILON)

	case CHAR_SET:
		for symbol := range token.value.(map[uint8]bool) {
			addTransition(start, end, symbol)
		}

	case GROUP:
		tokens := token.value.([]Token)
		start, end = tokenToNFA(&tokens[0])
		for i := 1; i < len(tokens); i++ {
			nextStart, nextEnd := tokenToNFA(&tokens[i])
			addTransition(end, nextStart, EPSILON)
			end = nextEnd
		}

	case REPEAT:
		tok := token.value.(RepeatData).token
		mn := token.value.(RepeatData).min
		mx := token.value.(RepeatData).max

		if mn == 0 {
			addTransition(start, end, EPSILON)
		}

		concatCount := 1
		if mx != INFINITY {
			concatCount = mx
		} else if mn != 0 {
			concatCount = mn
		}

		s, e := tokenToNFA(&tok)
		addTransition(start, s, EPSILON)

		for i := 2; i <= concatCount; i++ {
			nextStart, nextEnd := tokenToNFA(&tok)
			addTransition(e, nextStart, EPSILON)
			s = nextStart
			e = nextEnd

			if i > mn {
				addTransition(s, end, EPSILON)
			}
		}

		addTransition(e, end, EPSILON)

		if mx == INFINITY {
			addTransition(end, s, EPSILON)
		}

	default:
		exit("Error: unknown token type")
	}

	return start, end
}

func toNFA(tokens []Token) *State {
	startState, endState := tokenToNFA(&tokens[0])
	for i := 1; i < len(tokens); i++ {
		nextStart, nextEnd := tokenToNFA(&tokens[i])
		addTransition(endState, nextStart, EPSILON)
		endState = nextEnd
	}

	start := &State{
		start: true,
	}
	addTransition(start, startState, EPSILON)

	end := &State{
		end: true,
	}
	addTransition(endState, end, EPSILON)

	return start
}

func match(state *State, input string) bool {
	type CacheKey struct {
		state *State
		i     int
	}

	cache := make(map[CacheKey]bool)

	var matcher func(*State, string, int) bool

	matcher = func(state *State, input string, i int) bool {
		if result, ok := cache[CacheKey{state, i}]; ok {
			return result
		}

		if i == len(input) {
			if state.end {
				cache[CacheKey{state, i}] = true
				return true
			}

			for _, transition := range state.transitions {
				if transition.symbol != EPSILON {
					continue
				}
				for _, epsilonState := range transition.states {
					if matcher(epsilonState, input, i) {
						cache[CacheKey{state, i}] = true
						return true
					}
				}
			}

			cache[CacheKey{state, i}] = false
			return false
		}
		var c uint8 = input[i]

		for _, transition := range state.transitions {
			switch transition.symbol {
			case c:
				for _, nextState := range transition.states {
					if matcher(nextState, input, i+1) {
						cache[CacheKey{state, i}] = true
						return true
					}
				}
			case EPSILON:
				for _, epsilonState := range transition.states {
					if matcher(epsilonState, input, i) {
						cache[CacheKey{state, i}] = true
						return true
					}
				}
			}
		}

		cache[CacheKey{state, i}] = false
		return false
	}

	return matcher(state, input, 0)
}

func main() {
	tokens := parse(os.Args[1])
	startState := toNFA(tokens)
	fmt.Println(match(startState, os.Args[2]))
}
