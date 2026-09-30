package analyzer

import "sort"

type Entry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Metrics struct {
	Nu1 int     `json:"nu1"`
	Nu2 int     `json:"nu2"`
	N1  int     `json:"n1"`
	N2  int     `json:"n2"`
	Nu  int     `json:"nu"`
	N   int     `json:"N"`
	V   float64 `json:"V"`
}

type Result struct {
	// Старые поля оставлены для совместимости внутреннего API.
	// GUI больше их не отображает.
	Operators []Entry     `json:"operators"`
	Operands  []Entry     `json:"operands"`
	Metrics   Metrics     `json:"metrics"`
	Gilb      GilbMetrics `json:"gilb"`
}

func Analyze(source string) Result {
	// Halstead остаётся доступным для старого кода проекта,
	// но метрика Джилба считается независимо от его нормализации.
	tokens := Tokenize(source)
	normalizedTokens := Normalize(tokens)
	ops, opds := CountTokens(normalizedTokens)

	return Result{
		Operators: toSortedEntries(ops),
		Operands:  toSortedEntries(opds),
		Metrics:   CalculateHalstead(ops, opds),
		Gilb:      AnalyzeGilb(source),
	}
}

func CountTokens(tokens []Token) (map[string]int, map[string]int) {
	operators := make(map[string]int)
	operands := make(map[string]int)

	for _, t := range tokens {
		if t.Type == TokenOperator {
			operators[t.Value]++
		}
		if t.Type == TokenOperand {
			operands[t.Value]++
		}
	}
	return operators, operands
}

func toSortedEntries(m map[string]int) []Entry {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] != m[keys[j]] {
			return m[keys[i]] > m[keys[j]]
		}
		return keys[i] < keys[j]
	})
	out := make([]Entry, 0, len(keys))
	for _, k := range keys {
		out = append(out, Entry{Name: k, Count: m[k]})
	}
	return out
}
