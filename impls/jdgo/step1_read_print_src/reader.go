package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var intRegexp = regexp.MustCompile(`^-?[0-9]+$`)

// var floatRegexp = regexp.MustCompile(`-?[0-9][0-9.]*$`)
var tokenRegexp = regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" +
	`~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" +
	`,;)]*)`)

type MalType interface {
	TypeName() string
	Print() string
}

type List struct {
	items []MalType
}

func NewList() *List {
	return &List{}
}

func (list *List) TypeName() string { return "List" }
func (list *List) Print() string {
	var b strings.Builder
	b.WriteString("(")
	for _, v := range list.items {
		b.WriteString(v.Print())
		b.WriteString(" ")
	}
	r := b.String()
	r = strings.TrimRight(r, " ")
	return fmt.Sprintf("%s)", r)
}

func (list *List) Append(form MalType) {
	list.items = append(list.items, form)
}

type Vector struct {
	items []MalType
}

func NewVector() *Vector {
	return &Vector{}
}

func (vec *Vector) TypeName() string { return "Vector" }
func (vec *Vector) Print() string {
	vv := []string{}
	for _, v := range vec.items {
		vv = append(vv, v.Print())
	}
	return "[" + strings.Join(vv, " ") + "]"
}

func (vec *Vector) Append(form MalType) {
	vec.items = append(vec.items, form)
}

type HashMap struct {
	items map[String]MalType
}

func NewHashMap(forms []MalType) *HashMap {
	items := map[String]MalType{}

	for i := 0; i < len(forms); i += 2 {
		k := forms[i]
		v := forms[i+1]
		key, ok := k.(*String)
		if ok {
			items[*key] = v
		} else {
			fmt.Printf("WTF: %T %v", k, k)
		}
	}
	return &HashMap{items}
}

func (hm *HashMap) TypeName() string { return "HashMap" }
func (hm *HashMap) Print() string {
	str := []string{}
	for k, v := range hm.items {
		str = append(str, k.Print())
		str = append(str, v.Print())
	}
	return "{" + strings.Join(str, " ") + "}"
}

func (hm *HashMap) Set(key String, form MalType) {
	hm.items[key] = form
}

type Symbol struct {
	value string
}

func NewSymbol(value string) *Symbol {
	return &Symbol{value}
}

func (sym *Symbol) TypeName() string { return "Symbol" }
func (sym *Symbol) Print() string    { return sym.value }

type String struct {
	value   string
	keyword bool
}

func NewString(value string) *String {
	return &String{value, false}
}

func (str *String) TypeName() string { return "String" }
func (str *String) Print() string {
	if !str.keyword {
		return `"` + strings.Replace(
			strings.Replace(
				strings.Replace(str.value, `\`, `\\`, -1),
				`"`, `\"`, -1),
			"\n", `\n`, -1) + `"`
	}

	return fmt.Sprintf(":%s", str.value)
}

type Int struct {
	value int
}

func NewInt(s string) *Int {
	value, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return &Int{value: value}
}

func (i *Int) TypeName() string { return "Int" }
func (i *Int) Print() string    { return fmt.Sprintf("%d", i.value) }

type Float struct {
	value float64
}

func (f *Float) TypeName() string { return "Float" }
func (f *Float) Print() string    { return fmt.Sprintf("%f", f.value) }

type Boolean struct {
	value bool
}

func NewBoolean(value bool) *Boolean {
	return &Boolean{value}
}
func (b *Boolean) TypeName() string { return "Boolean" }
func (b *Boolean) Print() string {
	if b.value {
		return "true"
	}
	return "false"
}

type Nil struct {
}

func (n *Nil) TypeName() string { return "Nil" }
func (n *Nil) Print() string    { return "nil" }
func NewKeyword(value string) *String {
	return &String{strings.TrimLeft(value, ":"), true}
}

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
		l := &List{[]MalType{&Symbol{"quote"}, form}}
		return l, nil
	case "`":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := &List{[]MalType{&Symbol{"quasiquote"}, form}}
		return l, nil
	case "~":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := &List{[]MalType{&Symbol{"unquote"}, form}}
		return l, nil
	case "@":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := &List{[]MalType{&Symbol{"deref"}, form}}
		return l, nil
	case "~@":
		r.Next()
		form, err := r.ReadForm()
		if err != nil {
			return nil, err
		}
		l := &List{[]MalType{&Symbol{"splice-unquote"}, form}}
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

func PrintStr(ast MalType) string {
	return ast.Print()
}
