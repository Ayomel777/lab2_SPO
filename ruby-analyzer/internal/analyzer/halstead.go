package analyzer

import "math"

func CalculateHalstead(operators, operands map[string]int) Metrics {
	var m Metrics

	m.Nu1 = len(operators)
	m.Nu2 = len(operands)

	for _, c := range operators {
		m.N1 += c
	}
	for _, c := range operands {
		m.N2 += c
	}

	m.Nu = m.Nu1 + m.Nu2
	m.N = m.N1 + m.N2

	if m.Nu > 0 {
		m.V = float64(m.N) * math.Log2(float64(m.Nu))
	}

	return m
}
