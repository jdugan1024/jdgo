package main

import (
	"fmt"

	"github.com/chzyer/readline"
)

func READ(sexpr string) string {
	return sexpr
}
func EVAL(sexpr string) string {
	return sexpr
}
func PRINT(sexpr string) string {
	return sexpr
}

func rep(input string) string {
	return PRINT(READ(EVAL(input)))
}

func main() {
	rl, err := readline.New("user> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			// fmt.Println(err)
			break
		}

		fmt.Println(rep(input))
	}
}
