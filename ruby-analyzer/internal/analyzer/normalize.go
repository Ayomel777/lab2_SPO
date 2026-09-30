package analyzer

import "strings"

var builtinMethodOperators = map[string]bool{
	"puts": true, "print": true, "p": true,
	"each": true, "map": true, "select": true, "reject": true,
	"reduce": true, "inject": true, "each_with_index": true,
	"each_pair": true, "each_key": true, "each_value": true,
	"require": true, "require_relative": true, "load": true,
	"attr_accessor": true, "attr_reader": true, "attr_writer": true,
	"include": true, "extend": true, "prepend": true,
	"loop": true, "raise": true, "lambda": true, "proc": true,
	"gets": true, "chomp": true, "freeze": true,
	"push": true, "pop": true, "shift": true, "unshift": true,
	"length": true, "size": true, "first": true, "last": true,
	"keys": true, "values": true, "sort": true, "reverse": true,
	"upcase": true, "downcase": true, "capitalize": true, "strip": true,
	"split": true, "join": true, "to_s": true, "to_i": true, "to_f": true,
	"to_a": true, "to_h": true, "to_sym": true,
}

func Normalize(tokens []Token) []Token {
	tokens = markDefNames(tokens)
	tokens = mergeBlocks(tokens)
	tokens = mergeFunctionCalls(tokens)
	tokens = pairBrackets(tokens)
	tokens = markBuiltinMethods(tokens)
	return tokens
}

func markDefNames(tokens []Token) []Token {
	out := make([]Token, len(tokens))
	copy(out, tokens)
	for i := 0; i+1 < len(out); i++ {
		if out[i].Type == TokenOpenBlock && out[i].Value == "def" &&
			out[i+1].Type == TokenOperand {
			out[i+1].IsFuncName = true
		}
	}
	return out
}

func mergeBlocks(tokens []Token) []Token {
	type frame struct {
		opener string
		parts  []string
	}

	stack := make([]frame, 0, 16)
	out := make([]Token, 0, len(tokens))

	for _, t := range tokens {
		switch t.Type {
		case TokenOpenBlock:
			if t.Value == "do" && len(stack) > 0 {
				top := stack[len(stack)-1].opener
				if top == "while" || top == "until" || top == "for" {
					continue
				}
			}
			stack = append(stack, frame{opener: t.Value})

		case TokenBlockPart:
			if len(stack) > 0 {
				stack[len(stack)-1].parts = append(
					stack[len(stack)-1].parts, t.Value,
				)
			}

		case TokenCloseBlock:
			if len(stack) == 0 {
				out = append(out, Token{Value: t.Value, Type: TokenOperator})
				continue
			}
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			out = append(out, Token{
				Value: buildBlockName(f.opener, f.parts),
				Type:  TokenOperator,
			})

		default:
			out = append(out, t)
		}
	}

	for _, f := range stack {
		out = append(out, Token{Value: f.opener, Type: TokenOperator})
	}

	return out
}

func buildBlockName(opener string, parts []string) string {
	seen := make(map[string]bool, len(parts))
	uniq := make([]string, 0, len(parts))
	for _, p := range parts {
		if !seen[p] {
			seen[p] = true
			uniq = append(uniq, p)
		}
	}

	var sb strings.Builder
	sb.WriteString(opener)
	for _, p := range uniq {
		sb.WriteString("...")
		sb.WriteString(p)
	}
	sb.WriteString("...end")
	return sb.String()
}

func mergeFunctionCalls(tokens []Token) []Token {
	remove := make([]bool, len(tokens))
	convert := make([]string, len(tokens))

	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i].Type != TokenOperand {
			continue
		}
		if tokens[i].IsFuncName {
			continue
		}
		if !isSimpleIdentifier(tokens[i].Value) {
			continue
		}
		if tokens[i+1].Type != TokenOperator || tokens[i+1].Value != "(" {
			continue
		}

		depth := 0
		matched := -1
		for j := i + 1; j < len(tokens); j++ {
			if tokens[j].Type != TokenOperator {
				continue
			}
			if tokens[j].Value == "(" {
				depth++
			} else if tokens[j].Value == ")" {
				depth--
				if depth == 0 {
					matched = j
					break
				}
			}
		}
		if matched == -1 {
			continue
		}

		convert[i] = tokens[i].Value + "()"
		remove[i+1] = true
		remove[matched] = true
	}

	return applyReplacements(tokens, remove, convert)
}

func pairBrackets(tokens []Token) []Token {
	remove := make([]bool, len(tokens))
	convert := make([]string, len(tokens))

	type opener struct {
		idx int
		ch  string
	}
	stack := make([]opener, 0, 16)

	for i, t := range tokens {
		if t.Type != TokenOperator {
			continue
		}
		switch t.Value {
		case "(":
			stack = append(stack, opener{i, "("})
		case "[":
			stack = append(stack, opener{i, "["})
		case ")":
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.ch == "(" {
					convert[top.idx] = "()"
					remove[i] = true
					break
				}
			}
		case "]":
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.ch == "[" {
					convert[top.idx] = "[]"
					remove[i] = true
					break
				}
			}
		}
	}

	var pipes []int
	for i, t := range tokens {
		if t.Type == TokenOperator && t.Value == "|" {
			pipes = append(pipes, i)
		}
	}
	for k := 0; k+1 < len(pipes); k += 2 {
		convert[pipes[k]] = "||"
		remove[pipes[k+1]] = true
	}

	return applyReplacements(tokens, remove, convert)
}

func markBuiltinMethods(tokens []Token) []Token {
	out := make([]Token, len(tokens))
	copy(out, tokens)
	for i := range out {
		if out[i].Type == TokenOperand && builtinMethodOperators[out[i].Value] {
			out[i].Type = TokenOperator
		}
	}
	return out
}

func applyReplacements(tokens []Token, remove []bool, convert []string) []Token {
	out := make([]Token, 0, len(tokens))
	for i, t := range tokens {
		if remove[i] {
			continue
		}
		if convert[i] != "" {
			out = append(out, Token{Value: convert[i], Type: TokenOperator})
			continue
		}
		out = append(out, t)
	}
	return out
}

func isSimpleIdentifier(s string) bool {
	if s == "" {
		return false
	}
	last := len(s) - 1
	for i, r := range s {
		isLetter := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			r == '_'
		isDigit := r >= '0' && r <= '9'
		isSuffix := (r == '?' || r == '!') && i == last

		if i == 0 {
			if !isLetter {
				return false
			}
		} else if !isLetter && !isDigit && !isSuffix {
			return false
		}
	}
	return true
}
