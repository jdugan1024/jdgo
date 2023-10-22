package reader

import (
	"errors"
	"regexp"
	"strings"

	. "github.com/jdugan1024/jdgo/types"
)

var intRegexp = regexp.MustCompile(`^-?[0-9]+$`)

// var floatRegexp = regexp.MustCompile(`-?[0-9][0-9.]*$`)
var tokenRegexp = regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" +
	`~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" +
	`,;)]*)`)

type Reader struct {
	tokens   []string
	position int
}

func NewReader(tokens []string) *Reader {
	return &Reader{tokens: tokens, position: 0}
}

func (r *Reader) Peek() (string, error) {
	if r.position >= len(r.tokens) {
		return "", errors.New("EOF")
	}
	return r.tokens[r.position], nil
}

func (r *Reader) Next() (string, error) {
	t, err := r.Peek()
	if err != nil {
		return "", err
	}
	r.position++
	return t, nil
}

func (r *Reader) ReadForm() (MalType, error) {
	t, err := r.Peek()
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(t, ";") {
		r.Next()
	}
	if strings.HasPrefix(t, ",") {
		r.Next()
	}

	switch t {
	case ")":
		return nil, errors.New("unexpected )")
	case "(":
		return r.ReadList()
	case "]":
		return nil, errors.New("unexpected ]")
	case "[":
		return r.ReadVector()
	case "}":
		return nil, errors.New("unexpected }")
	case "{":
		return r.ReadHashMap()
	case "'":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := NewList(NewSymbol("quote"), form)
		return l, nil
	case "`":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := NewList(NewSymbol("quasiquote"), form)
		return l, nil
	case "~":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := NewList(NewSymbol("unquote"), form)
		return l, nil
	case "@":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := NewList(NewSymbol("deref"), form)
		return l, nil
	case "~@":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := NewList(NewSymbol("splice-unquote"), form)
		return l, nil
	default:
		return r.ReadAtom()
	}
}

func (r *Reader) ReadAtom() (MalType, error) {
	t, err := r.Next()
	if err != nil {
		return nil, err
	}
	if intRegexp.MatchString(t) {
		return NewInt(t), nil
	}
	if strings.HasPrefix(t, ":") {
		return NewKeyword(t), nil
	}
	if match, _ := regexp.MatchString(`^"(?:\\.|[^\\"])*"$`, t); match {
		str := (t)[1 : len(t)-1]
		s := strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(str, `\\`, "\u029e", -1),
					`\"`, `"`, -1),
				`\n`, "\n", -1),
			"\u029e", "\\", -1)
		return NewString(s), nil
	}
	if strings.HasPrefix(t, `"`) {
		return nil, errors.New("EOF")
	}

	switch t {
	case "true":
		return NewBoolean(true), nil
	case "false":
		return NewBoolean(false), nil
	case "nil":
		return &Nil{}, nil
	}

	return NewSymbol(t), nil
}

func (r *Reader) ReadList() (*List, error) {
	_, err := r.Next()
	if err != nil {
		return nil, err
	}
	l := NewList()

	for {
		t, err := r.Peek()

		if err != nil {
			return nil, err
		}
		if t == ")" {
			r.Next()
			break
		}

		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l.Append(form)
	}

	return l, nil
}

func (r *Reader) ReadVector() (*Vector, error) {
	_, err := r.Next()
	if err != nil {
		return nil, err
	}
	v := NewVector()

	for {
		t, err := r.Peek()

		if err != nil {
			return nil, err
		}

		if t == "]" {
			_, err := r.Next()
			if err != nil {
				return nil, err
			}
			break
		}

		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		v.Append(form)
	}

	return v, nil
}

func (r *Reader) ReadHashMap() (*HashMap, error) {

	_, err := r.Next()
	if err != nil {
		return nil, err
	}

	forms := []MalType{}

	for {
		t, err := r.Peek()

		if err != nil {
			return nil, err
		}

		if t == "}" {
			_, err := r.Next()
			if err != nil {
				return nil, err
			}
			break
		}

		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		forms = append(forms, form)
	}

	if len(forms)%2 != 0 {
		return nil, errors.New("uneven number of entries in hashmap")
	}

	return NewHashMap(forms), nil
}

func Tokenize(input string) []string {
	rawTokens := tokenRegexp.FindAllStringSubmatch(input, -1)
	tokens := []string{}

	for _, t := range rawTokens {
		tokens = append(tokens, t[1])
	}

	return tokens
}
