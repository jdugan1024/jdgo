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
func EVAL(ast MalType, env *Env) (MalType, error) {
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
	return apply(l, env)
}
func PRINT(ast MalType) string {
	return PrintStr(ast)
}

func apply(l *List, env *Env) (MalType, error) {
	head, err := l.First()
	if err != nil {
		return nil, errors.New("can't apply an empty list")
	}
	symbol, ok := head.(*Symbol)
	if ok {
		value := symbol.Print()
		switch value {
		case "def!":
			rest, err := l.Rest()
			if err != nil {
				return nil, errors.New("missing args for def!")
			}
			key, ok := rest[0].(*Symbol)
			if !ok {
				return nil, fmt.Errorf("env key is not a symbol: %s", rest[0].Print())
			}
			value, err := EVAL(rest[1], env)
			if err != nil {
				return nil, fmt.Errorf("unable to eval arg for def! %s %s", key.Print(), rest[1].Print())
			}
			env.Set(key, value)
			return value, nil
		case "let*":
			rest, err := l.Rest()
			if err != nil {
				return nil, err
			}

			bindingsObj := rest[0]
			switch bindings := bindingsObj.(type) {
			case *List:
				if bindings.Length()%2 != 0 {
					return nil, fmt.Errorf("let* bindings has an odd number of entries: %s", bindings.Print())
				}

				newEnv := NewEnv(env)
				bindings.BindEnv(newEnv, EVAL)

				r, err := EVAL(rest[1], newEnv)

				if err != nil {
					return nil, err
				}
				return r, nil
			case *Vector:
				if bindings.Length()%2 != 0 {
					return nil, fmt.Errorf("let* bindings has an odd number of entries: %s", bindings.Print())
				}

				newEnv := NewEnv(env)
				bindings.BindEnv(newEnv, EVAL)

				r, err := EVAL(rest[1], newEnv)

				if err != nil {
					return nil, err
				}
				return r, nil
			default:
				return nil, fmt.Errorf("bindings is not a list or a vector: %s", bindingsObj.Print())
			}

		}
	}
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

var replEnv = NewEnv(nil)

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

func eval_ast(ast MalType, env *Env) (MalType, error) {
	switch v := ast.(type) {
	case *Symbol:
		return env.Get(v)
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

	replEnv.Set(NewSymbol("+"), NewFunction("+", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[0].Print())
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[1].Print())
		}

		return NewIntFromInt(a.AsInt() + b.AsInt()), nil
	}))

	replEnv.Set(NewSymbol("-"), NewFunction("-", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}

		return NewIntFromInt(a.AsInt() - b.AsInt()), nil
	}))
	replEnv.Set(NewSymbol("*"), NewFunction("*", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[0].Print())
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, fmt.Errorf("argument is not an Int: %s", args[1].Print())
		}

		return NewIntFromInt(a.AsInt() * b.AsInt()), nil
	}))
	replEnv.Set(NewSymbol("/"), NewFunction("/", func(args ...MalType) (MalType, error) {
		a, ok := args[0].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}
		b, ok := args[1].(*Int)
		if !ok {
			return nil, errors.New("argument is not an Int")
		}

		return NewIntFromInt(a.AsInt() / b.AsInt()), nil
	}))
	replEnv.Set(NewSymbol("print"), NewFunction("print", func(args ...MalType) (MalType, error) {
		var s = ""
		for _, v := range args {
			s += v.Print()
		}
		fmt.Println(s)
		return &Nil{}, nil
	}))
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
