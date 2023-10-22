package main

import (
	"errors"
	"fmt"

	"github.com/chzyer/readline"

	. "github.com/jdugan1024/jdgo/printer"
	. "github.com/jdugan1024/jdgo/reader"
	. "github.com/jdugan1024/jdgo/types"
)

func READ(input string) (MalType, error) {
	reader := NewReader(Tokenize(input))
	ast, err := reader.ReadForm()
	if err != nil {
		return nil, err
	}
	return ast, nil
}
func EVAL(ast MalType, env Env) (MalType, error) {
	switch ast.(type) {
	case *List:
	default:
		return eval_ast(ast, env)
	}
	l, ok := ast.(*List)
	if !ok {
		return nil, errors.New("not a list")
	}
	if l.Length() == 0 {
		return l, nil
	}
	return applyList(l, env)
}
func PRINT(ast MalType) string {
	return PrintStr(ast)
}

func applyList(l *List, env Env) (MalType, error) {
	e, err := eval_ast(l, env)
	if err != nil {
		return nil, err
	}
	el, ok := e.(*List)
	if !ok {
		return nil, err
	}
	l0, err := el.First()
	if err != nil {
		return nil, err
	}
	f, ok := l0.(*Function)
	if !ok {
		return nil, fmt.Errorf("%s is not a function", l0.Print())
	}
	args, err := el.Rest()
	if err != nil {
		return nil, err
	}
	return f.Eval(args...)
}

var replEnv = Env{
	"+": NewFunction("+", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[0].Print())
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[1].Print())
		}

		return NewIntFromInt(a.AsInt() + b.AsInt()), nil
	}),
	"-": NewFunction("-", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}

		return NewIntFromInt(a.AsInt() - b.AsInt()), nil
	}),
	"*": NewFunction("*", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[0].Print())
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[1].Print())
		}

		return NewIntFromInt(a.AsInt() * b.AsInt()), nil
	}),
	"/": NewFunction("/", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}

		return NewIntFromInt(a.AsInt() / b.AsInt()), nil
	}),
}

func rep(input string) (string, error) {
	ast, err := READ(input)
	if err != nil {
		return "", err
	}
	ev, err := EVAL(ast, replEnv)
	if err != nil {
		return "", err
	}
	return PRINT(ev), nil
}

func eval_ast(ast MalType, env Env) (MalType, error) {
	switch v := ast.(type) {
	case *Symbol:
		name := v.Print()
		r, ok := env[name]
		if !ok {
			return nil, fmt.Errorf("unknown symbol: %s", name)
		}
		return r, nil
	case *List:
		r, err := v.Map(EVAL, env)
		if err != nil {
			return nil, err
		}
		return r, nil
	case *Vector:
		r, err := v.Map(EVAL, env)
		if err != nil {
			return nil, err
		}
		return r, nil
	case *HashMap:
		fmt.Printf(">>> %s\n", v.Print())
		r, err := v.Map(EVAL, env)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return ast, nil
	}

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
