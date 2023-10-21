package main

import (
	"fmt"

	"github.com/chzyer/readline"
)

func READ(input string) (MalType, error) {
	reader := NewReader(Tokenize(input))
	ast, err := reader.ReadForm()
	if err != nil {
		return nil, err
	}
	return ast, nil
}
func EVAL(ast MalType) (MalType, error) {
	return ast, nil
}
func PRINT(ast MalType) string {
	return PrintStr(ast)
}

func rep(input string) (string, error) {
	ast, err := READ(input)
	if err != nil {
		return "", err
	}
	ev, err := EVAL(ast)
	if err != nil {
		return "", err
	}
	return PRINT(ev), nil
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

		r, err := rep(input)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)
	}
}
