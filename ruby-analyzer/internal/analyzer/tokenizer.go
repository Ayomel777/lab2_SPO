package analyzer

import (
	"strings"
	"unicode"
)

func isIdentifierStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentifierPart(r rune) bool {
	return unicode.IsLetter(r) ||
		unicode.IsDigit(r) ||
		r == '_'
}

func startsWith(runes []rune, position int, value string) bool {
	valueRunes := []rune(value)
	if position+len(valueRunes) > len(runes) {
		return false
	}
	for i := 0; i < len(valueRunes); i++ {
		if runes[position+i] != valueRunes[i] {
			return false
		}
	}
	return true
}

func Tokenize(source string) []Token {
	runes := []rune(source)
	tokens := make([]Token, 0)

	i := 0

	for i < len(runes) {

		if unicode.IsSpace(runes[i]) {
			i++
			continue
		}

		// Комментарий
		if runes[i] == '#' {
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			continue
		}

		// Строки
		if runes[i] == '"' || runes[i] == '\'' {
			quote := runes[i]
			i++

			var builder strings.Builder
			builder.WriteRune(quote)

			for i < len(runes) {
				if runes[i] == '\\' && i+1 < len(runes) {
					builder.WriteRune(runes[i])
					builder.WriteRune(runes[i+1])
					i += 2
					continue
				}
				builder.WriteRune(runes[i])
				if runes[i] == quote {
					i++
					break
				}
				i++
			}

			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand})
			continue
		}

		// @ivar / @@cvar
		if runes[i] == '@' {
			var builder strings.Builder
			builder.WriteRune(runes[i])
			i++
			if i < len(runes) && runes[i] == '@' {
				builder.WriteRune(runes[i])
				i++
			}
			for i < len(runes) && isIdentifierPart(runes[i]) {
				builder.WriteRune(runes[i])
				i++
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand})
			continue
		}

		// $global
		if runes[i] == '$' {
			var builder strings.Builder
			builder.WriteRune(runes[i])
			i++
			for i < len(runes) && isIdentifierPart(runes[i]) {
				builder.WriteRune(runes[i])
				i++
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand})
			continue
		}

		// Числа
		if unicode.IsDigit(runes[i]) {
			var builder strings.Builder
			hasDot := false
			for i < len(runes) {
				if unicode.IsDigit(runes[i]) || runes[i] == '_' {
					builder.WriteRune(runes[i])
					i++
					continue
				}
				if runes[i] == '.' && !hasDot && i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
					hasDot = true
					builder.WriteRune(runes[i])
					i++
					continue
				}
				break
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand})
			continue
		}

		// :symbol
		if runes[i] == ':' && i+1 < len(runes) && runes[i+1] != ':' && isIdentifierStart(runes[i+1]) {
			var builder strings.Builder
			builder.WriteRune(':')
			i++
			for i < len(runes) && isIdentifierPart(runes[i]) {
				builder.WriteRune(runes[i])
				i++
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand})
			continue
		}

		// Идентификатор / ключевое слово
		if isIdentifierStart(runes[i]) {
			var builder strings.Builder
			for i < len(runes) && isIdentifierPart(runes[i]) {
				builder.WriteRune(runes[i])
				i++
			}
			if i < len(runes) && (runes[i] == '?' || runes[i] == '!') {
				builder.WriteRune(runes[i])
				i++
			}
			word := builder.String()

			if blockOpeners[word] {
				tokens = append(tokens, Token{Value: word, Type: TokenOpenBlock})
				continue
			}
			if word == "end" {
				tokens = append(tokens, Token{Value: word, Type: TokenCloseBlock})
				continue
			}
			if blockParts[word] {
				tokens = append(tokens, Token{Value: word, Type: TokenBlockPart})
				continue
			}
			if keywordOperators[word] {
				tokens = append(tokens, Token{Value: word, Type: TokenOperator})
				continue
			}
			tokens = append(tokens, Token{Value: word, Type: TokenOperand})
			continue
		}

		// Скобки
		if runes[i] == '(' {
			tokens = append(tokens, Token{Value: "(", Type: TokenOperator})
			i++
			continue
		}
		if runes[i] == ')' {
			tokens = append(tokens, Token{Value: ")", Type: TokenOperator})
			i++
			continue
		}
		if runes[i] == '[' {
			tokens = append(tokens, Token{Value: "[", Type: TokenOperator})
			i++
			continue
		}
		if runes[i] == ']' {
			tokens = append(tokens, Token{Value: "]", Type: TokenOperator})
			i++
			continue
		}
		if runes[i] == '{' {
			tokens = append(tokens, Token{Value: "{", Type: TokenOperator})
			i++
			continue
		}
		if runes[i] == '}' {
			tokens = append(tokens, Token{Value: "}", Type: TokenOperator})
			i++
			continue
		}

		// Многосимвольные операторы
		found := false
		for _, operator := range multiCharOperators {
			if startsWith(runes, i, operator) {
				tokens = append(tokens, Token{Value: operator, Type: TokenOperator})
				i += len([]rune(operator))
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Односимвольные операторы
		current := string(runes[i])
		if symbolOperators[current] {
			tokens = append(tokens, Token{Value: current, Type: TokenOperator})
			i++
			continue
		}

		i++
	}

	return tokens
}
