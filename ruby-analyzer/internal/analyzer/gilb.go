package analyzer

import "sort"

// GilbConstruct описывает управляющую конструкцию Ruby,
// участвующую в расчёте метрики Джилба.
type GilbConstruct struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// GilbOperator описывает оператор Ruby из спецификации языка.
type GilbOperator struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// GilbMetrics содержит данные для метрики Джилба и перечень
// конструкций/операторов Ruby, которые были найдены в программе.
type GilbMetrics struct {
	CL         int             `json:"cl"`
	Relative   float64         `json:"relative"`
	CLI        int             `json:"cli"`
	Operators  int             `json:"operators"`
	Constructs []GilbConstruct `json:"constructs"`
	RubyOps    []GilbOperator  `json:"rubyOps"`
}

type gilbFrame struct {
	kind string

	// Для CASE.
	caseBranches     int
	caseBranchActive bool
	outerDepth       int
	caseHasElse      bool
}

// Ruby operators. Пунктуация языка (скобки, запятые, точки с запятой,
// фигурные/квадратные скобки как разделители и т.п.) сюда не входит.
// Методы Ruby также не считаются операторами: foo(), puts(), each() и т.п.
var rubyOperators = map[string]bool{
	// assignment
	"=": true, "+=": true, "-=": true, "*=": true, "/=": true,
	"%=": true, "**=": true, "&=": true, "|=": true, "^=": true,
	"<<=": true, ">>=": true, "&&=": true, "||=": true,

	// arithmetic / unary
	"+": true, "-": true, "*": true, "/": true, "%": true, "**": true,

	// comparison
	"==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
	"<=>": true, "===": true,

	// pattern matching
	"=~": true, "!~": true,

	// logical
	"&&": true, "||": true, "!": true,
	"and": true, "or": true, "not": true,

	// bitwise / shift
	"&": true, "|": true, "^": true, "~": true,
	"<<": true, ">>": true,

	// range
	"..": true, "...": true,
}

// AnalyzeGilb рассчитывает метрики Джилба непосредственно по исходному Ruby-коду.
// Halstead-нормализация здесь не используется.
func AnalyzeGilb(source string) GilbMetrics {
	tokens := Tokenize(source)

	cl := 0
	currentDepth := 0
	maxDepth := 0
	activeCaseDepth := 0

	stack := make([]gilbFrame, 0)
	constructCounts := make(map[string]int)
	operatorCounts := make(map[string]int)

	// Считаем только операторы, определённые как операторы Ruby,
	// плюс управляющие конструкции, которые являются операторами программы
	// при определении N.
	for _, token := range tokens {
		if token.Type == TokenOperator && rubyOperators[token.Value] {
			operatorCounts[token.Value]++
		}
	}

	for _, token := range tokens {
		switch token.Type {
		case TokenOpenBlock:
			// В Ruby do после while/until/for не открывает второй
			// управляющий блок.
			if token.Value == "do" && len(stack) > 0 {
				top := stack[len(stack)-1].kind
				if top == "while" || top == "until" || top == "for" {
					continue
				}
			}

			frame := gilbFrame{kind: token.Value}

			switch token.Value {
			case "if", "unless", "while", "until", "for":
				cl++
				currentDepth++
				stack = append(stack, frame)
				constructCounts[token.Value]++

				effectiveDepth := currentDepth + activeCaseDepth
				if effectiveDepth > maxDepth {
					maxDepth = effectiveDepth
				}

			case "case":
				frame.outerDepth = currentDepth
				stack = append(stack, frame)
				constructCounts["case"]++
			}

		case TokenBlockPart:
			switch token.Value {
			case "when":
				caseIndex := findNearestCase(stack)
				if caseIndex == -1 {
					continue
				}

				stack[caseIndex].caseBranches++
				constructCounts["when"]++

				if !stack[caseIndex].caseBranchActive {
					stack[caseIndex].caseBranchActive = true
					activeCaseDepth++

					effectiveDepth := currentDepth + activeCaseDepth
					if effectiveDepth > maxDepth {
						maxDepth = effectiveDepth
					}
				}

			case "elsif":
				cl++
				constructCounts["elsif"]++

				effectiveDepth := currentDepth + activeCaseDepth
				if effectiveDepth > maxDepth {
					maxDepth = effectiveDepth
				}

			case "else":
				// else не является отдельным условием.
				// Важен только ELSE, относящийся непосредственно к CASE.
				// Если сверху находится IF/WHILE, это ELSE внутренней
				// конструкции и CASE не получает ветвь по умолчанию.
				if len(stack) > 0 && stack[len(stack)-1].kind == "case" {
					stack[len(stack)-1].caseHasElse = true
				}
			}

		case TokenCloseBlock:
			if len(stack) == 0 {
				continue
			}

			index := len(stack) - 1
			frame := stack[index]
			stack = stack[:index]

			switch frame.kind {
			case "if", "unless", "while", "until", "for":
				if currentDepth > 0 {
					currentDepth--
				}

			case "case":
				n := frame.caseBranches

				// CASE с n when без else/default даёт n условий.
				// При наличии else/default последняя ветвь является
				// ветвью по умолчанию и условий становится n-1.
				caseCL := n
				if frame.caseHasElse && caseCL > 0 {
					caseCL--
				}
				cl += caseCL

				// Для CLI CASE раскрывается в цепочку условий.
				// Максимальная глубина цепочки определяется количеством
				// проверяемых альтернатив.
				caseDepth := caseCL
				if caseDepth > 0 {
					caseDepth--
				}
				absoluteDepth := frame.outerDepth + caseDepth
				if absoluteDepth > maxDepth {
					maxDepth = absoluteDepth
				}

				if frame.caseBranchActive && activeCaseDepth > 0 {
					activeCaseDepth--
				}
			}
		}
	}

	// N = управляющие конструкции + операторы Ruby.
	// case учитывается один раз в N, хотя его вклад в CL раскрывается
	// по числу альтернатив.
	operators := 0
	for name, count := range constructCounts {
		// when — часть CASE, а не самостоятельный оператор программы.
		// Поэтому показываем его в списке конструкций, но не включаем в N.
		if name != "when" {
			operators += count
		}
	}
	for _, count := range operatorCounts {
		operators += count
	}

	relative := 0.0
	if operators > 0 {
		relative = float64(cl) / float64(operators)
	}

	constructs := make([]GilbConstruct, 0, len(constructCounts))
	for name, count := range constructCounts {
		constructs = append(constructs, GilbConstruct{Name: name, Count: count})
	}
	sort.Slice(constructs, func(i, j int) bool {
		return constructs[i].Name < constructs[j].Name
	})

	rubyOps := make([]GilbOperator, 0, len(operatorCounts))
	for name, count := range operatorCounts {
		rubyOps = append(rubyOps, GilbOperator{Name: name, Count: count})
	}
	sort.Slice(rubyOps, func(i, j int) bool {
		if rubyOps[i].Count != rubyOps[j].Count {
			return rubyOps[i].Count > rubyOps[j].Count
		}
		return rubyOps[i].Name < rubyOps[j].Name
	})

	return GilbMetrics{
		CL:         cl,
		Relative:   relative,
		CLI:        maxDepth,
		Operators:  operators,
		Constructs: constructs,
		RubyOps:    rubyOps,
	}
}

func findNearestCase(stack []gilbFrame) int {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].kind == "case" {
			return i
		}
	}
	return -1
}
