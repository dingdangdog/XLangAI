package captcha

import (
	"fmt"
	"math/rand"
)

type mathOp string

const (
	opAdd mathOp = "+"
	opSub mathOp = "-"
	opMul mathOp = "*"
	opDiv mathOp = "/"
)

func pickOperator() mathOp {
	ops := []mathOp{opAdd, opSub, opMul, opDiv}
	return ops[rand.Intn(len(ops))]
}

func digit0to9() int { return rand.Intn(10) }

func generateNumbers(op mathOp) (a, b int) {
	a, b = digit0to9(), digit0to9()
	if op != opDiv {
		return a, b
	}
	b = rand.Intn(9) + 1 // 1..9
	quotient := rand.Intn(10)
	a = b * quotient
	tries := 0
	for a > 9 && tries < 20 {
		b = rand.Intn(9) + 1
		q := rand.Intn(10)
		a = b * q
		tries++
	}
	if a > 9 {
		b = rand.Intn(9) + 1
		a = b
	}
	return a, b
}

func computeAnswer(a, b int, op mathOp) float64 {
	switch op {
	case opAdd:
		return float64(a + b)
	case opSub:
		return float64(a - b)
	case opMul:
		return float64(a * b)
	case opDiv:
		if b == 0 {
			return 0
		}
		return float64(a) / float64(b)
	default:
		return 0
	}
}

// GenerateQuestion 生成四则运算题（参与数 0-9，除法整除）。
func GenerateQuestion() (question string, answer float64) {
	op := pickOperator()
	a, b := generateNumbers(op)
	if op == opDiv && (b == 0 || a%b != 0) {
		a, b = generateNumbers(op)
	}
	answer = computeAnswer(a, b, op)
	question = fmt.Sprintf("%d %s %d = ?", a, op, b)
	return question, answer
}
