package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MalType interface {
	TypeName() string
	Print() string
}
type Env struct {
	outer *Env
	items map[string]MalType
}

func NewEnv(outer *Env) *Env {
	return &Env{outer, map[string]MalType{}}
}

func (env *Env) Set(k *Symbol, v MalType) {
	env.items[k.value] = v
}

func (env *Env) Find(k *Symbol) (MalType, error) {
	workingEnv := env
	for {
		v, ok := workingEnv.items[k.value]
		if ok {
			return v, nil
		}
		workingEnv = workingEnv.outer
		if workingEnv == nil {
			break
		}
	}
	return nil, fmt.Errorf("'%s' not found", k.value)
}

func (env *Env) Get(k *Symbol) (MalType, error) {
	return env.Find(k)
}

type List struct {
	items []MalType
}

func NewList(items ...MalType) *List {
	return &List{items}
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
func (list *List) Map(f func(arg MalType, env *Env) (MalType, error), env *Env) (MalType, error) {
	r := []MalType{}

	for _, v := range list.items {
		item, err := f(v, env)
		if err != nil {
			return nil, err
		}
		r = append(r, item)
	}

	return &List{r}, nil
}
func (list *List) Length() int { return len(list.items) }
func (list *List) First() (MalType, error) {
	if list.Length() == 0 {
		return nil, errors.New("can't take first of empty list")
	}
	return list.items[0], nil
}
func (list *List) Rest() ([]MalType, error) {
	if list.Length() == 0 {
		return nil, errors.New("can't take first of empty list")
	}
	return list.items[1:], nil
}
func (list *List) BindEnv(env *Env, eval func(MalType, *Env) (MalType, error)) error {
	for i := 0; i < len(list.items); i += 2 {
		keyObj := list.items[i]
		key, ok := keyObj.(*Symbol)
		if !ok {
			return fmt.Errorf("attempting to bind to a non symbol: %s", keyObj.Print())
		}
		val, err := eval(list.items[i+1], env)
		if err != nil {
			return err
		}
		env.Set(key, val)
	}
	return nil
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

func (vec *Vector) Map(f func(arg MalType, env *Env) (MalType, error), env *Env) (MalType, error) {
	r := []MalType{}

	for _, v := range vec.items {
		item, err := f(v, env)
		if err != nil {
			return nil, err
		}
		r = append(r, item)
	}

	return &Vector{r}, nil
}
func (vec *Vector) Length() int { return len(vec.items) }
func (vec *Vector) First() (MalType, error) {
	if vec.Length() == 0 {
		return nil, errors.New("can't take first of empty list")
	}
	return vec.items[0], nil
}
func (vec *Vector) Rest() ([]MalType, error) {
	if vec.Length() == 0 {
		return nil, errors.New("can't take first of empty list")
	}
	return vec.items[1:], nil
}
func (vec *Vector) BindEnv(env *Env, eval func(MalType, *Env) (MalType, error)) error {
	for i := 0; i < len(vec.items); i += 2 {
		keyObj := vec.items[i]
		key, ok := keyObj.(*Symbol)
		if !ok {
			return fmt.Errorf("attempting to bind to a non symbol: %s", keyObj.Print())
		}
		val, err := eval(vec.items[i+1], env)
		if err != nil {
			return err
		}
		env.Set(key, val)
	}
	return nil
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
func (hm *HashMap) Map(f func(arg MalType, env *Env) (MalType, error), env *Env) (MalType, error) {
	r := map[String]MalType{}

	for k, v := range hm.items {
		item, err := f(v, env)
		if err != nil {
			return nil, err
		}
		r[k] = item
	}

	return &HashMap{r}, nil
}
func (hm *HashMap) Length() int { return len(hm.items) }

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
func NewIntFromInt(i int) *Int {
	return &Int{value: i}
}

func (i *Int) TypeName() string { return "Int" }
func (i *Int) Print() string    { return fmt.Sprintf("%d", i.value) }
func (i *Int) AsInt() int       { return i.value }

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

type Function struct {
	name string
	f    func(...MalType) (MalType, error)
}

func NewFunction(name string, f func(...MalType) (MalType, error)) *Function {
	return &Function{name, f}
}

func (f *Function) TypeName() string { return "Function" }
func (f *Function) Print() string    { return f.name }
func (f *Function) Eval(args ...MalType) (MalType, error) {
	return f.f(args...)
}
