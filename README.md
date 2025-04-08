# gorgx

A regular expression engine in Go implemented using [Finite-state machine](https://en.wikipedia.org/wiki/Finite-state_machine).
this engine support most of the regex syntax like:

- Literals
 - Matches exactly what you typed `hello` matches "hello"
- Character Classes 
 - The bracket syntax for example `[abc]` matches a, b or c and `[a-z]` matches any char from a to z
- Quantifiers
 - `*`  = 0 or more 
 - `+` = 1 or more 
 - `?` = 0 or 1 
 - `{n}`  = Exactly n times
 - `{n,}` = n or more
 - `{n,m}`  = at least n at most m
- Alternation (OR)
 - `cat|dog` matches either "cat" or "dog"
- Grouping
 - `([0-9]{4})/([0-9]{2})/([0-9]{2})` matches  "2025/04/08" 

### Usage
after cloning this repo you have to have Golang on you system and then
```bash
go build
./gorgx "PATTERN" "STRING"
```
if the string matches the pattern it will print true, false otherwise

### Implementation Overview
it consists of 3 phases:
 
- Parsing
 - in this phase we transform the regex string into tokens that has meaning
 - if the regex string is invalid we exit the program early with the error message
- Building the state machine
 - in this phase we take the tokens form the parsing phase and make an NFA (non-deterministic finite automata) out of it
 - each token has its unique automata and after generating them we concat them in order
- Matching
 - it is just running the input string against the NFA if the input string is matches we print true otherwise false
