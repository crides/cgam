package main

import (// <<<<
    "bufio"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "math"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "os/user"
    "sort"
    "strconv"
    "strings"
    "syscall"
    "time"
)// >>>>
// TODO The math.Remainder function is a bit weird
// TODO The alphabet operators
// The memo for recursive functions (`y`)// <<<<
type Memo struct {
    block   *Block
    arity   int
    cache   map[interface{}]interface{}
}

func NewMemo(b *Block, n int) *Memo {
    return &Memo{b, n, make(map[interface{}]interface{})}
}

func (m * Memo) Set(entry []interface{}, val interface{}) {
    if len(entry) != m.arity {
        panic("`y` function arity and entry length doesn't match!")
    }
    m.cache[&entry] = val
}

//def set(self, l, o, n):
//    if n == 0:
//        if self.n == 1:
//            self.m[repr(l[0])] = o
//        else:
//            self.m[repr(l)] = o
//        return
//    ol = o
//    for i in range(len(ol)):
//        l.append(i)
//        self.set(l, ol[i], n - 1)
//        l.pop()

func (m * Memo) Get(entry []interface{}) (interface{}, bool) {
    if len(entry) != m.arity {
        panic("`y` function arity and entry length doesn't match!")
    }
    item, ok := m.cache[&entry]
    return item, ok
}

func (m * Memo) Run(env * Environ) {
    var args []interface{}
    for i := 0; i < m.arity; i ++ {
        args = append(args, env.Pop())
    }
    Reverse(args)

    if res, cached := m.Get(args); cached {
        env.Push(res)
        return
    }
    for _, a := range args {
        env.Push(a)
    }
    m.block.Run(env)
    m.Set(args, env.stack.Get(0))
}
// >>>>
// The stack// <<<<
func wraps(as ...interface{}) *Stack {       // Wrap as a stack
    return NewStack(as)
}

func wrapa(as ...interface{}) []interface{} {       // Wrap as a list
    return as
}

func wrap(a interface{}) interface{} {              // Just convert type
    return a
}

type Stack []interface{}

func NewStack(x []interface{}) *Stack {
    new_stack := new(Stack)
    if x != nil {
        *new_stack = x
    }
    return new_stack
}

func (x * Stack) Push(a interface{}) {
    *x = append(wrapa(a), *x...)
}

func (x * Stack) Pusha(as *Stack) {
    *x = append(*as, *x...)
}

func (x * Stack) Pop() (a interface{}) {
    a, *x = (*x)[0], (*x)[1:]
    // TODO reset env.marks if needed
    return
}

func (x * Stack) Get(ind int) interface{} {
    return (*x)[ind]
}

func (x * Stack) Get1() interface{} {
    return (*x)[0]
}

func (x * Stack) Get2() (interface{}, interface{}) {
    return (*x)[0], (*x)[1]
}

func (x * Stack) Get3() (interface{}, interface{}, interface{}) {
    return (*x)[0], (*x)[1], (*x)[2]
}

func (x * Stack) Cut(a, b int) *Stack {
    return NewStack((*x)[a:b])
}

func (x * Stack) Dump() {
    fmt.Println("--- Top of Stack ---")
    for _, i := range *x {
        //s := to_s(i)
        //fmt.Printf("%s | %v\n", s, i)
        fmt.Println(to_v(i))
    }
    fmt.Println("--- Bottom of Stack ---")
}

func (x * Stack) Clear() {
    *x = []interface{}{}
}

func (x * Stack) Size() int {
    return len(*x)
}

func (x * Stack) Contents() []interface{} {
    return []interface{}(*x)
}

func (x * Stack) Switch(a, b int) {
    (*x)[a], (*x)[b] = (*x)[b], (*x)[a]
}

func (x * Stack) Reverse() {    // Reverse in-place
    for i, j := 0, x.Size() - 1; i <= j; i, j = i + 1, j - 1 {
        x.Switch(i, j)
    }
}

func Reverse(arr []interface{}) {    // Reverse in-place
    for i, j := 0, len(arr) - 1; i <= j; i, j = i + 1, j - 1 {
        arr[i], arr[j] = arr[j], arr[i]
    }
}
// >>>>
// The environment// <<<<
type Environ struct {
    stack       *Stack
    marks       []int
    memo        *Memo
    vars        []interface{}
    longVars    map[string]interface{}
    args        []string
}

func NewEnviron() *Environ {
    env := &Environ{NewStack(nil),
                    make([]int, 0),
                    nil,
                    make([]interface{}, 0),
                    make(map[string]interface{}),
                    os.Args[1:]}
    env.InitVars()
    return env
}

func (env * Environ) SetMemo(m * Memo) {
    env.memo = m
}

func (env * Environ) GetMemo() *Memo {
    return env.memo
}

func (env * Environ) Pop() (a interface{}) {    // Wraps the Stack.pop() method to resize the env.marks.
    a = env.stack.Pop()
    size := env.stack.Size()
    for i, m := range env.marks {
        if m > size {
            env.marks[i] = size
        }
    }
    return
}

func (env * Environ) Pop2() (a, b interface{}) {
    b, a = env.Pop(), env.Pop()
    return
}

func (env * Environ) Pop3() (a, b, c interface{}) {
    c, b, a = env.Pop(), env.Pop(), env.Pop()
    return
}

func (env * Environ) Push(a interface{}) {
    env.stack.Push(a)
}

func (env * Environ) Pusha(as *Stack) {
    env.stack.Pusha(as)
}

func (env * Environ) Mark() {
    env.marks = append(env.marks, env.stack.Size())
}

func (env * Environ) PopMark() {
    mark := 0
    size := len(env.marks)
    if size > 0 {
        mark, env.marks = env.marks[size - 1], env.marks[:size - 1]
    }
    end := env.stack.Size() - mark
    wrapped := NewStack((*env.stack)[:end])
    wrapped.Reverse()       // For compat
    *env.stack = append(wrapa([]interface{}(*wrapped)), (*env.stack)[end:]...)
}

// Variables
func (env * Environ) GetVar(s string) interface{} {
    if len(s) == 1 {
        if c := s[0]; is_upper(rune(c)) {
            return env.vars[c - 'A']
        }
        panic("Invalid variable name")
    }
    if a, ok := env.longVars[s]; ok {
        return a
    }
    panic("Variable not defined!")
}

func (env * Environ) SetVar(s string, a interface{}) {
    if len(s) == 1 {
        if c := s[0]; is_upper(rune(c)) {
            env.vars[c - 'A'] = a
            return
        }
        panic("Invalid variable name")
    }
    env.longVars[s] = a
}

func (env * Environ) InitVars() {
    for i := 0; i <= 10; i ++ {
        env.vars = append(env.vars, 10 + i)
    }
    env.vars = append(env.vars, wrapa([]rune{}, []rune{}, []rune{'\n'}, []rune{}, math.Pi, []rune{}, []rune{}, []rune{' '}, 0, 0, 0, -1, 1, 2, 3)...)
    //                                 L   M    N    O   P        Q   R    S   T  U  V  W   X  Y  Z
}

func (env * Environ) ResetVars() {
    env.vars = make([]interface{}, 0)
    //env.longVars = make(map[string]interface{})
}
// >>>>
// Type predicates//<<<<
func typeof(a interface{}) string {
    switch a.(type) {
    case []rune, string:
        return "string"
    case []interface{}:
        for _, i := range a.([]interface{}) {
            if ! is_c(i) {
                return "list"
            }
        }
        return "string"
    case rune:
        return "char"
    case int:
        return "int"
    case float64:
        return "double"
    case *Block:
        return "block"
    }
    return "unknown"
}

// Full type (Cgam type and Go type) of the object. Used in debug.
func fulltype(a interface{}) string {
    return fmt.Sprintf("%s (%T)", typeof(a), a)
}

func type_eq(a interface{}, typ rune) bool {
    // All the typenames thru out the program mean the same thing:// <<<<
    // (including the type predicates `is_x` and conversions `to_x` functions)
    //
    // a: all
    // b: block
    // c: char (or rune, actually :))
    // d: double
    // i: int
    // l: list (or []interface{})
    // n: num
    // p: num + char ('p' = 'n' + 'c' - 'a')
    // s: string
    // v: value (anything other than block)
    //
    // The only exception being that to_v means `to value string` (or repr() in Python)
// >>>>
    switch typ {
    case 'a':
        if is_n(a) || is_l(a) || is_b(a) || is_c(a) {
            return true
        }
        panic("Invalid type!: " + fulltype(a))
    case 'b':
        return is_b(a)
    case 'c':
        return is_c(a)
    case 'd':
        return is_d(a)
    case 'i':
        return is_i(a)
    case 'l':
        return is_l(a)
    case 'n':
        return is_n(a)
    case 'p':
        return is_n(a) || is_c(a)
    case 's':
        return is_s(a)
    case 'v':
        return is_n(a) || is_l(a) || is_c(a)
    }
    panic("Invalid typename!: `" + string(typ) + "`")
}

func any_of(a, b interface{}, typ string) bool {
    return typeof(a) == typ || typeof(b) == typ
}

func is_n(a interface{}) bool {
    return is_i(a) || is_d(a)
}

func is_i(a interface{}) bool {
    return typeof(a) == "int"
}

func is_d(a interface{}) bool {
    return typeof(a) == "double"
}

func is_c(a interface{}) bool {
    return typeof(a) == "char"
}

func is_l(a interface{}) bool {
    return typeof(a) == "list" || is_s(a)
}

func is_s(a interface{}) bool {
    return typeof(a) == "string"
}

func is_b(a interface{}) bool {
    return typeof(a) == "block"
}

func any_double(a, b interface{}) bool {
    return any_of(a, b, "double")
}
//>>>>
// Type conversions// <<<<
func not_convertible(a interface{}, typ string) {
    panic(fmt.Sprintf("Cannot convert %s to %s!", fulltype(a), typ))
}

func to_d(a interface{}) float64 {
    switch b := a.(type) {
    case int:
        return float64(b)
    case rune:
        return float64(int(b))
    case float64:
        return b
    case string:
        r, _ := strconv.ParseFloat(b, 64)
        return r
    case []rune:
        r, _ := strconv.ParseFloat(string(b), 64)
        return r
    default:
        not_convertible(a, "double")
    }
    return 0
}

func to_i(a interface{}) int {
    switch b := a.(type) {
    case int:
        return b
    case bool:
        if b {
            return 1
        }
        return 0
    case rune:
        return int(b)
    case float64:
        return int(b)
    case string:
        r, _ := strconv.ParseInt(b, 10, 0)
        return int(r)
    case []rune:
        r, _ := strconv.ParseInt(string(b), 10, 0)
        return int(r)
    default:
        not_convertible(a, "int")
    }
    return 0
}

func to_l(a interface{}) []interface{} {
    switch b := a.(type) {
    case []interface{}:
        return b
    case []rune:
        as := make([]interface{}, 0)
        for _, i := range b {
            as = append(as, wrap(i))
        }
        return as
    default:
        not_convertible(a, "list")
    }
    return nil
}

func to_b(a interface{}) *Block {
    switch b := a.(type) {
    case *Block:
        return b
    default:
        not_convertible(a, "Block")
    }
    return nil
}

func to_c(a interface{}) rune {
    switch b := a.(type) {
    case rune:
        return b
    case int:
        return rune(b)
    case float64:
        return rune(int(b))
    default:
        not_convertible(a, "char")
    }
    return 0
}

func to_s(a interface{}) string {
    switch b := a.(type) {
    case string:
        return b
    case []rune:
        return string(b)
    case rune:
        return string(b)
    case []interface{}:
        s := ""
        for _, i := range b {
            s += to_s(i)
        }
        return s
    case fmt.Stringer:
        return b.String()
    case error:
        return b.Error()
    default:
        return fmt.Sprint(b)
    }
}

func to_v(a interface{}) string {       // To value string
    switch {
    case is_n(a):
        return fmt.Sprint(a)
    case is_s(a):
        return fmt.Sprintf("%q", to_s(a))
    case is_c(a):
        return "'" + to_s(a)
    case is_l(a):
        s := "["
        for _, i := range to_l(a) {
            s += to_v(i) + " "
        }
        if len(s) == 1 {
            return "[]"
        }
        return s[:len(s) - 1] + "]"
    case is_b(a):
        return to_b(a).String()
    default:
        panic(fmt.Sprintf("Invalid type %s!", fulltype(a)))
    }
}

func to_bool(a interface{}) bool {
    switch b := a.(type) {
    case int, rune:
        return b != 0
    case float64:
        return b != 0
    case []interface{}:
        return len(b) != 0
    case []rune:
        return len(b) != 0
    default:
        not_convertible(a, "bool")
    }
    return false
}
//>>>>
// Block//<<<<
type Block struct {
    Ops         []*Op
    Offsets     [][2]int    // LineNum & Offset
    withbrace   bool        // Whether the block has braces when parsing

    parser      *Parser     // Link the parser for source code reference
}

func NewBlock(ops []*Op, offsets [][2]int, wb bool, parser *Parser) *Block {
    if len(ops) == len(offsets) {
        return &Block{ops, offsets, wb, parser}
    }
    panic("Sizes not matched!")
}

func (b * Block) Run(env * Environ) {
    for i, op := range b.Ops {
        if i == 0 {
            defer func() {
                if err := recover(); err != nil {
                    offsets := b.Offsets[i]
                    lnum := offsets[0]
                    column := offsets[1]
                    if err != SIGNAL {  // Error is a native error; prints header
                        fmt.Println("\x1b[31mERROR\x1b[0m: Runtime:", err.(string))
                    }
                    fmt.Printf(
`  at %s line %d:
    %s
    %s^
`,
                    // Both line number and column should start from 1.
                    b.parser.GetSrc(), lnum + 1,
                    strings.Split(string(b.parser.content), "\n")[lnum],
                    strings.Repeat(" ", column - 1))
                    panic(SIGNAL)
                }
            }()
        }
        op.Run(env)
    }
}

func (b * Block) String() string {
    s := []byte{'{'}
    for _, op := range b.Ops {
        s = append(s, append([]byte(op.String()), ' ')...)
    }
    if s[len(s) - 1] == ' ' {
        s[len(s) - 1] = '}'
    } else if b.withbrace {
        s = append(s, '}')
    }
    if ! b.withbrace {
        return string(s[1:len(s) - 1])
    }
    return string(s)
}

var SIGNAL string = "!#\ufdd0"

type RuntimeError struct {
    LnNum   int
    Offset  int

    Msg     string
}

func NewRuntimeError(ln, offs int, msg string) *RuntimeError {
    return &RuntimeError{ln, offs, msg}
}

func (e * RuntimeError) Error() string {
    return e.Msg
}
//>>>>
type Runner interface {// <<<<
    Run(*Environ)
    String() string
}

// A work around for the Op, Op1, Op2 and Op3 classes
func (op * Op) HasArity(arity int) bool {
    for _, tf := range op.Funcs {
        if len(tf.Sign) == arity {
            return true
        }
    }
    return false
}// >>>>
// Operator//<<<<
type TypedFunc struct {
    Sign    string        // Signature
    Option  byte          // * * * Wrapped | Switch_start Switch_end
    Func    func(env * Environ, x * Stack) *Stack
}

type Op struct {
    Name    string
    Funcs   []TypedFunc
}

func (op * Op) String() string {
    return op.Name
}

var OPS map[string]*Op = make(map[string]*Op)

func addOp(op * Op) {
    new_name := op.Name
    if _, ok := OPS[new_name]; ok {
        panic("Duplicate: " + new_name + "!")
    }
    OPS[new_name] = op
}

func findOp(name string) *Op {
    op, ok := OPS[name]
    if ! ok {
        panic("Unknown operator: `" + name + "`!")
    }
    return op
}

func op_push(a interface{}) *Op {
    return &Op{to_v(a),
        []TypedFunc{
            {"",
            0x10,
            func(env * Environ, x * Stack) *Stack {
                return wraps(a)
            }},
        }}
}

func op_pushVar(s string) *Op {
    return &Op{":" + s,
    []TypedFunc{
        {"",
        0x00,
        func(env * Environ, x * Stack) *Stack {
            a := env.GetVar(s)
            if ! is_b(a) {
                x.Push(a)
            } else {
                to_b(a).Run(env)
            }
            return nil
        }},
    }}
}

func op_setVar(s string) *Op {
    return &Op{"." + s,
    []TypedFunc{
        {"", 0x00,
        func(env * Environ, x * Stack) *Stack {
            env.SetVar(s, x.Get(0))
            return nil
        }},
    }}
}

func op_vector(r Runner) *Op {
    return &Op{":" + r.String(),
    []TypedFunc{
        {"ll", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            al, bl := to_l(a), to_l(b)
            as, bs := len(al), len(bl)
            size := int(math.Max(float64(as), float64(bs)))

            env.Mark()
            for i := 0; i < size; i ++ {
                if i < as {
                    env.Push(al[i])
                    if i < bs {
                        env.Push(bl[i])
                        r.Run(env)
                    }
                } else {
                    env.Push(bl[i])
                }
            }
            env.PopMark()
            return nil
        }},
    }}
}

func op_map(r Runner) *Op {
    return &Op{"." + r.String(),
    []TypedFunc{
        {"l", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a := env.Pop()
            al := to_l(a)
            env.Mark()
            for _, i := range al {
                env.Push(i)
                r.Run(env)
            }
            env.PopMark()
            return nil
        }},
    }}
}

func op_fold(op * Op) *Op {
    return &Op{":" + op.String(),
    []TypedFunc{
        {"l", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a := env.Pop()
            al := to_l(a)
            env.Push(al[0])
            for i := 1; i < len(al); i ++ {
                env.Push(al[i])
                op.Run(env)
            }
            return nil
        }},
    }}
}

func op_for(name string) *Op {
    return &Op{"f" + name,
    []TypedFunc{
        {"vb", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            bb := to_b(b)
            if is_l(a) {
                for _, i := range to_l(a) {
                    env.SetVar(name, i)
                    bb.Run(env)
                }
            } else {
                for i := 0; i < to_i(a); i ++ {
                    env.SetVar(name, i)
                    bb.Run(env)
                }
            }
            return nil
        }},
    }}
}

func op_map2(r Runner) *Op {
    return &Op{"f" + r.String(),
    []TypedFunc{
        {"la", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            for _, i := range to_l(a) {
                env.Push(i)
                env.Push(b)
                r.Run(env)
            }
            return nil
        }},

        {"vl", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            for _, i := range to_l(b) {
                env.Push(a)
                env.Push(i)
                r.Run(env)
            }
            return nil
        }},
    }}
}

func op_e10(a interface{}) *Op {
    if ! is_n(a) {
        panic("Argument to exp is not number!")
    }
    return &Op{"e" + to_s(a),
    []TypedFunc{
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            n := x.Get1()
            res := to_d(n) * math.Pow(10, to_d(a))
            if res == to_d(to_i(res)) && ! any_double(a, n) {
                return wraps(to_i(res))
            }
            return wraps(res)
        }},
    }}
}

func MatchSign(tf TypedFunc, x * Stack) int {
    // Return value:
    // -1 not enough arguments
    // 0: not match
    // 1: matched
    // 2: special match (match if switched)
    arity := len(tf.Sign)
    if x.Size() < arity {
        return -1
    }
    for i, typename := range tf.Sign {
        if ! type_eq(x.Get(arity - i - 1), typename) {
            goto try_switch
        }
    }
    // Success!
    return 1

try_switch:
    switch_opts := tf.Option & 0x0F
    if switch_opts > 0 {
        new_sign := []byte(tf.Sign)
        sw_s, sw_e := switch_opts & 0x0C, switch_opts & 0x03
        var new_tf TypedFunc
        new_sign[sw_s], new_sign[sw_e] = new_sign[sw_e], new_sign[sw_s]
        new_tf.Sign = string(new_sign)
        if MatchSign(new_tf, x) > 0 {
            // A special match
            return 2
        }
    }
    // Not matched
    return 0
}

func (op * Op) Run(env * Environ) {
    // Call functions// <<<<
    // For simplicity, the top of the stack is args[0]
    // But the arguments passed to Op.Run() is in the reverse direction
    //           -----------
    // args[0]    Environ top   c
    //           -----------
    // args[1]       ...        b
    //           -----------
    // args[2]       ...        a
    //           -----------
    // args[...]     ...
    //           -----------
    // args[-1]    Bottom
// >>>>
    func_no, func_arity := 0, 0
    var match int       // The match flag
    neas := 0           // Number of "Not enough arguments"

    // Try to match signature
    for i, tf := range op.Funcs {
        match = MatchSign(tf, env.stack)
        if match <= 0 {      // Not matched
            if match == -1 {
                neas ++
            }
        } else {
            // Signature matched
            func_no = i
            func_arity = len(tf.Sign)
            break
        }
    }
    if match > 0 {
        // Prepare the arguements; switch if needed
        op_func := op.Funcs[func_no]
        wrapped := (op_func.Option & 0x10) > 0
        switch_opts := op_func.Option & 0x0F
        sw_s, sw_e := int(switch_opts & 0x0C), int(switch_opts & 0x03)

        var raw_args []interface{}
        var args *Stack
        if wrapped {
            for i := 0; i < func_arity; i ++ {
                raw_args = append(raw_args, env.Pop())
            }
            args = NewStack(raw_args)
            args.Reverse()
            if match == 2 {      // Only do a switch if it's a special match
                args.Switch(sw_s, sw_e)
            }
        } else {
            args = env.stack
            if match == 2 {
                args.Switch(func_arity - sw_s - 1, func_arity - sw_e - 1)
            }
        }

        // Calling the function
        result := op_func.Func(env, args)
        if wrapped {
            env.Pusha(result)
        }
    } else {
        if neas == len(op.Funcs) {
            panic("Too few things on stack to call `" + op.String() + "`!")
        }
        // loop not end normally
        report_size := env.stack.Size() // How many arguments should report
        last_arity := len(op.Funcs[len(op.Funcs) - 1].Sign)
        if last_arity < report_size {
            report_size = last_arity
        }
        types := make([]string, report_size)
        for i := 0; i < report_size; i ++ {
            types[i] = typeof(env.stack.Get(report_size - i - 1))
        }
        panic(fmt.Sprintf("%s `%s` not implemented!", strings.Join(types, " "), op.Name))
    }
}
//>>>>
// Parsing//<<<<
type Parser struct {
    source  string

    content []rune
    ptr     int
    size    int
    err     error

    LnNum   int
    Offset  int
}

func NewParser(src, code string) *Parser {
    content := []rune(code)
    return &Parser{src,
                   content, 0, len(content), nil,
                   0, 0}
}

func (p * Parser) GetSrc() string {
    return p.source
}

func (p * Parser) GetOffset() [2]int {
    return [2]int{p.LnNum, p.Offset}
}

func (p * Parser) Read() (c rune, e error) {
    if p.ptr >= p.size {
        p.err = errors.New("EOF")
    }
    if p.err != nil {
        return 0, p.err
    }
    c = p.content[p.ptr]
    p.ptr ++
    p.Offset ++
    if c == '\n' {
        p.LnNum ++
        p.Offset = 0
    }
    return
}

func (p * Parser) Unread() error {
    if p.ptr == 0 {
        p.err = errors.New("Start reached!")
    }
    if p.err != nil {
        return p.err
    }
    p.Offset --
    p.ptr --
    return nil
}

func (p * Parser) UnreadN(n int) {
    for i := 0; i < n; i ++ {
        if p.Unread() != nil {
            panic(fmt.Sprintf("UnreadN(%d) reached start!", n))
        }
    }
}

func parse(code *Parser, withbrace bool) *Block {
    ops := make([]*Op, 0)
    offsets := make([][2]int, 0)

    for c, err := code.Read(); ; c, err = code.Read() {
        if err != nil {
            if withbrace {
                panic("Unfinished block")
            }
            return NewBlock(ops, offsets, withbrace, code)
        }

        switch c {
        case ' ', '\t', '\n':
        case '}':
            if withbrace {
                return NewBlock(ops, offsets, withbrace, code)
            }
            panic("Unexpected `}`")
        case ';':       // A line comment
            for c, err := code.Read(); err == nil && c != '\n'; c, err = code.Read() { }
        default:
            offsets = append(offsets, code.GetOffset())
            code.Unread()
            op := parseOp(code)
            ops = append(ops, op)
        }
    }
}

func parseNumber(code *Parser) interface{} {
    var num_str []rune       // String repr for the result number
    float, nega := false, false

    for c, err := code.Read(); err == nil; c, err = code.Read() {
        switch {
        case is_digit(c):
            num_str = append(num_str, c)
        case c == '.':
            if float {
                code.Unread()
                goto end_parse
            }
            float = true
            num_str = append(num_str, c)
        case c == '-':
            if nega {
                code.Unread()
                goto end_parse
            }
            nega = true
            num_str = append(num_str, c)
        default:
            code.Unread()
            goto end_parse
        }
    }

end_parse:
    if float {
        num_f, _ := strconv.ParseFloat(string(num_str), 64)
        return num_f
    }
    num_i, _ := strconv.ParseInt(string(num_str), 10, 0)
    return int(num_i)
}

func parseOp(code *Parser) *Op {
    char, err := code.Read()
    if err != nil {
        panic("Expects operator!")
    }

    if is_digit(char) {
        code.Unread()
        return op_push(parseNumber(code))
    }

    if is_upper(char) {
        return op_pushVar(string(char))
    }

    switch char {
    case '{':       // Block
        return op_push(parse(code, true))
    case '"':       // String
        var str []rune
        for char, err := code.Read(); err == nil; char, err = code.Read() {
            if char == '"' {
                return op_push(str)
            }
            if char == '\\' {
                c, err := code.Read()
                if err != nil {
                    panic("Unfinished string!")
                }

                if c == '"' || c == '\\' {
                    char = c
                }
            }
            str = append(str, char)
        }
        panic("Unfinished string!")

    case '\'':      // Char
        actual_char, err := code.Read()
        if err != nil {
            panic("Unfinished char!")
        }
        return op_push(actual_char)

    case '-':
        if next, _ := code.Read(); is_digit(next) {     // Check if next is digit
            code.Unread()
            code.Unread()
            return op_push(parseNumber(code))
        }
        code.Unread()       // Spit out the "possible digit"
        return findOp("-")

    case ':':
        next, _ := code.Read()
        if next == '{' {
            return op_vector(parse(code, true))
        } else if is_upper(next) {
            long_var_name := string(next)
            for next, _ = code.Read(); is_varchar(next); next, _ = code.Read() {
                long_var_name += string(next)
            }
            code.Unread()
            return op_pushVar(long_var_name)
        } else {
            code.Unread()
            op := parseOp(code)
            if op.HasArity(2) {
                return op_vector(op)
            }
            panic("Invalid operator after `.`: " + op.String())
        }

    case '.':
        next, _ := code.Read()
        if next == '{' {
            return op_map(parse(code, true))
        } else if is_upper(next) {
            var_name := string(next)
            for next, _ = code.Read(); is_varchar(next); next, _ = code.Read() {
                var_name += string(next)
            }
            code.Unread()
            return op_setVar(var_name)
        }
        code.Unread()
        op := parseOp(code)
        if op.HasArity(1) {
            return op_map(op)
        }
        if op.HasArity(2) {
            return op_fold(op)
        }
        panic("Invalid operator after `:`: " + op.String())

    case 'f':
        next, _ := code.Read()
        if next == '{' {
            return op_map2(parse(code, true))
        } else if is_upper(next) {
            return op_for(string(next))
        }
        code.Unread()
        op := parseOp(code)
        if op.HasArity(2) {
            return op_map2(op)
        }
        panic("Invalid operator after `f`: " + op.String())

    case 'e':
        next, _ := code.Read()
        if is_digit(next) || next == '-' || next == '.' {
            code.Unread()
            return op_e10(parseNumber(code))
        }
        return findOp(string(char) + string(next))

    case 'm':
        next, _ := code.Read()
        if is_digit(next) || next == '-' || next == '.' {
            code.Unread()
            return findOp("-")
        }
        return findOp(string(char) + string(next))

    case 'o', 'r', 'x':       // The `os`, `regex`, `xfer` modules
        next, _ := code.Read()
        return findOp(string(char) + string(next))

    default:
        return findOp(string(char))
    }
}
//>>>>
// Functions //<<<<
func InitFuncs() {
    // Note:// <<<<
    // 1. The return value in the wrapped functions must be wraps()'ed;
    //
    // 2. For non-wrapping functions, the `x` is actually the stack in `env`;
    //
    // 3. For non-wrapping functions, the return value can be `nil`, because it is actually
    //    discarded;
    //
    // 4. Sometimes a condition would be put in front of the others because checking the type
    //    first would substract that situation from another situation. In this kind of
    //    situation, often the latter situation's signature is hard to represent.
    //
    // 5. The order of the operators is the order of the corespondent Unicode code-points.// >>>>

    // Symbol operators// <<<<
    addOp(&Op{"!",// <<<<
    []TypedFunc{
        // Logical NOT
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(! to_bool(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"#",// <<<<
    []TypedFunc{
        // Numeric power
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            res := math.Pow(to_d(a), to_d(b))
            if res == to_d(to_i(res)) && ! any_double(a, b) {
                return wraps(to_i(res))
            }
            return wraps(res)
        }},

        // Find index
        {"lv", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al := to_l(a)
            bl := wrapa()
            if is_l(b) {
                bl = to_l(b)
            } else {
                bl = wrapa(b)
            }
            return wraps(find(al, bl))
        }},

        // Find index that satisfy block
        {"lb", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            al, bb := to_l(a), to_b(b)
            for i := 0; i < len(al); i ++ {
                x.Push(al[i])
                bb.Run(env)
                if to_bool(env.Pop()) {
                    x.Push(i)
                    return nil
                }
            }
            x.Push(-1)
            return nil
        }},
    }})// >>>>
    addOp(&Op{"$",// <<<<
    []TypedFunc{
        // Copy from stack
        {"n", 0x00,
        func(env * Environ, x * Stack) *Stack {
            ind := to_i(env.Pop())
            if ind > 0 {
                x.Push(x.Get(ind))
            } else {
                x.Push(x.Get(x.Size() + ind))
            }
            return nil
        }},

        // Simple sort
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            l := x.Get1()
            s := NewSorter(to_l(l), I)
            sort.Stable(s)
            return wraps(s.arr)
        }},

        // Sort by key
        {"lb", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            al, bb := to_l(a), to_b(b)

            s := NewSorter(al, func(a interface{}) interface{} {
                x.Push(a)
                bb.Run(env)
                return env.Pop()
            })
            sort.Stable(s)
            x.Push(s.arr)
            return nil
        }},
    }})// >>>>
    addOp(&Op{"%",// <<<<
    []TypedFunc{
        // Modulo
        // TODO The math.Remainder function is a bit weird
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(math.Remainder(to_d(a), to_d(b)))
            }
            return wraps(to_i(a) % to_i(b))
        }},

        // Every nth item
        {"ln", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            as := len(al)
            res := wrapa()
            var rev bool
            switch {
            case bi == 0:
                panic("Invalid step!")
            case bi < 0:
                rev = true
                bi = -bi
            case bi > 0:
            }

            var ind int
            for i := 0; i < as; i += bi {
                if rev {
                    ind = as - 1 - i
                } else {
                    ind = i
                }
                res = append(res, al[ind])
            }
            return wraps(res)
        }},

        // Split (no empty parts)
        {"lv", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if is_l(b) {
                return wraps(split(to_l(a), to_l(b), false))
            } else {
                return wraps(split(to_l(a), wrapa(b), false))
            }
        }},

        // Foreach (wraps in a list)
        {"bv", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            env.Mark()
            ab := to_b(a)
            if is_c(b) || is_n(b) {
                bi := to_i(b)
                for i := 0; i < bi; i ++ {
                    x.Push(i)
                    ab.Run(env)
                }
            } else {
                bl := to_l(b)
                for _, i := range bl {
                    x.Push(i)
                    ab.Run(env)
                }
            }
            env.PopMark()
            return nil
        }},
    }})// >>>>
    addOp(&Op{"&",// <<<<
    []TypedFunc{
        // Bitwise AND
        {"ii", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) & to_i(b))
        }},

        {"ic", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(to_i(a) & to_i(b)))
        }},

        // Set intersection
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bl := wrapa(), wrapa()
            res := wrapa()
            if is_l(a) {
                al = to_l(a)
            } else {
                al = wrapa(a)
            }
            if is_l(b) {
                bl = to_l(b)
            } else {
                bl = wrapa(b)
            }
            for _, i := range al {
                if find(bl, wrapa(i)) >= 0 && find(res, wrapa(i)) == -1 {
                    res = append(res, i)
                }
            }
            return wraps(res)
        }},

        // If-then
        {"vb", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            if to_bool(a) {
                to_b(b).Run(env)
            }
            return nil
        }},
    }})// >>>>
    addOp(&Op{"(",// <<<<
    []TypedFunc{
        // Decrement
        {"p", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            var res interface{}
            switch {
            case is_c(a):
                res = to_c(a) - 1
            case is_i(a):
                res = to_i(a) - 1
            case is_d(a):
                res = to_d(a) - 1
            }
            return wraps(res)
        }},

        // Uncons from left
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            al := to_l(a)
            o, al := al[0], al[1:]
            return wraps(o, al)
        }},
    }})// >>>>
    addOp(&Op{")",// <<<<
    []TypedFunc{
        // Increment
        {"p", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            var res interface{}
            switch {
            case is_c(a):
                res = to_c(a) + 1
            case is_i(a):
                res = to_i(a) + 1
            case is_d(a):
                res = to_d(a) + 1
            }
            return wraps(res)
        }},

        // Uncons from right
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            al := to_l(a)
            as := len(al)
            o, al := al[as - 1], al[:as - 1]
            return wraps(o, al)
        }},
    }})// >>>>
    addOp(&Op{"*",// <<<<
    []TypedFunc{
        // Numeric multiplication
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) * to_d(b))
            }
            return wraps(to_i(a) * to_i(b))
        }},

        // Repeat value
        {"vi", 0x11,
        func(env * Environ, x * Stack) *Stack {
            s := []interface{}{}
            a, b := x.Get2()
            for i := 0; i < int(to_i(b)); i ++ {
                if is_l(a) {
                    s = append(s, to_l(a)...)
                } else {
                    s = append(s, wrapa(a)...)
                }
            }
            return wraps(s)
        }},

        // Repeat block execution
        {"bi", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            for i := 0; int(i) < to_i(b); i ++ {
                to_b(a).Run(env)
            }
            return nil
        }},

        // Join
        {"lv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            res := wrapa()
            bl := wrapa()

            if ! is_l(b) {
                bl = wrapa(b)
            } else {
                bl = to_l(b)
            }

            for i, o := range to_l(a) {
                if i > 0 {
                    res = append(res, to_l(bl)...)
                }
                if is_l(o) {
                    res = append(res, to_l(o)...)
                } else {
                    res = append(res, o)
                }
            }
            return wraps(res)
        }},

        {"cl", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            res := wrapa()

            for i, o := range to_l(b) {
                if i > 0 {
                    res = append(res, a)
                }
                if is_l(o) {
                    res = append(res, to_l(o)...)
                } else {
                    res = append(res, o)
                }
            }
            return wraps(res)
        }},

        // Fold / Reduce
        {"lb", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            al, bb := to_l(a), to_b(b)
            if len(al) > 0 {
                x.Push(al[0])
                for i := 1; i < len(al); i ++ {
                    x.Push(al[i])
                    bb.Run(env)
                }
            }
            return nil
        }},
    }})
// >>>>
    addOp(&Op{"+",// <<<<
    []TypedFunc{
        // Numeric addition
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) + to_d(b))
            }
            return wraps(to_i(a) + to_i(b))
        }},

        // Character concatenation
        {"cc", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps([]rune{a.(rune), b.(rune)})
        }},

        // Character incrementation (-> Char)
        {"cn", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(a) + to_c(b))
        }},

        // List concat
        {"ll", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(append(to_l(a), to_l(b)...))
        }},

        // List append
        {"la", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(append(to_l(a), b))
        }},

        // List append
        {"al", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(append(wrapa(a), to_l(b)...))
        }},
    }})
// >>>>
    addOp(&Op{",",// <<<<
    []TypedFunc{
        // Range(stop)
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            n := to_i(x.Get1())
            if n < 0 {
                panic("Invalid size for range!")
            }
            res := wrapa()
            for i := 0; i < n; i ++ {
                res = append(res, i)
            }
            return wraps(res)
        }},

        // Range for chars
        {"c", 0x10,
        func(env * Environ, x * Stack) *Stack {
            n := to_i(x.Get1())
            res := wrapa()
            for i := 0; i < n; i ++ {
                res = append(res, to_c(i))
            }
            return wraps(res)
        }},

        // Length
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            l := to_l(x.Get1())
            return wraps(len(l))
        }},

        // Filter
        {"lb", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            al, bb := to_l(a), to_b(b)
            res := wrapa()
            for _, i := range al {
                x.Push(i)
                bb.Run(env)
                if to_bool(env.Pop()) {
                    res = append(res, i)
                }
            }
            x.Push(res)
            return nil
        }},

        {"nb", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            ai, bb := to_i(a), to_b(b)
            res := wrapa()
            for i := 0; i < ai; i ++ {
                x.Push(i)
                bb.Run(env)
                if to_bool(env.Pop()) {
                    res = append(res, i)
                }
            }
            x.Push(res)
            return nil
        }},
    }})// >>>>
    addOp(&Op{"-",// <<<<
    []TypedFunc{
        // Numeric minus
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) - to_d(b))
            }
            return wraps(to_i(a) - to_i(b))
        }},

        // Character decrementation (-> Char)
        {"cn", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(a) - to_c(b))
        }},

        // Character difference (-> int)
        {"cc", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) - to_i(b))
        }},

        // Remove from list
        {"la", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al := to_l(a)
            res := wrapa()
            if is_l(b) {
                for _, i := range al {
                    if find(to_l(b), wrapa(i)) == -1 {
                        res = append(res, i)
                    }
                }
            } else {
                for _, i := range to_l(a) {
                    if i == b {
                        continue
                    }
                    res = append(res, i)
                }
            }
            return wraps(res)
        }},

        {"pl", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            for _, i := range to_l(b) {
                if i == a {
                    return wraps(wrapa())
                }
            }
            return wraps(wrapa(a))
        }},
    }})
// >>>>
    addOp(&Op{"/",// <<<<
    []TypedFunc{
        // Numberic division
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) / to_d(b))
            }
            return wraps(to_i(a) / to_i(b))
        }},

        // Split by length
        {"ln", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            n := to_i(b)
            if n <= 0 {
                panic("Invalid size for spliting!")
            }
            al := to_l(a)
            m := len(al)
            res := wrapa()
            end := 0
            for i := 0; i < m; i += n {
                if i + n < m {
                    end = i + n
                } else {
                    end = m
                }
                res = append(res, al[i:end])
            }
            return wraps(res)
        }},

        // Split by sep
        {"lv", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if is_l(b) {
                return wraps(split(to_l(a), to_l(b), true))
            } else {
                return wraps(split(to_l(a), wrapa(b), true))
            }
        }},

        // Foreach
        {"bv", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            ab := to_b(a)
            if is_c(b) || is_n(b) {
                bn := to_i(b)
                for i := 0; i < bn; i ++ {
                    x.Push(i)
                    ab.Run(env)
                }
            } else {
                bl := to_l(b)
                for _, i := range bl {
                    x.Push(i)
                    ab.Run(env)
                }
            }
            return nil
        }},
    }})
// >>>>
    addOp(&Op{"`",  // From `;` to ```// <<<<
    []TypedFunc{
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"<",// <<<<
    []TypedFunc{
        // Slice before
        {"li", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            return wraps(al[:adjust_ind(bi, len(al))])
        }},

        // Compare (Less than)
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(comp(a, b) < 0))
        }},
    }})// >>>>
    addOp(&Op{"=",// <<<<
    []TypedFunc{
        // Get from list
        {"li", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            return wraps(al[adjust_indm(bi, len(al))])
        }},

        // Compare (Equals)
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(comp(a, b) == 0))
        }},

        // Find value satisfy condition
        {"lb", 0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            al, bb := to_l(a), to_b(b)
            for _, i := range al {
                x.Push(i)
                bb.Run(env)
                if to_bool(env.Pop()) {
                    x.Push(i)
                    return nil
                }
            }
            return nil
        }},
    }})// >>>>
    addOp(&Op{">",// <<<<
    []TypedFunc{
        // Slice after
        {"li", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            return wraps(al[adjust_ind(bi, len(al)):])
        }},

        // Compare (Greater than)
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(comp(a, b) > 0))
        }},
    }})// >>>>
    addOp(&Op{"?",// <<<<
    []TypedFunc{
        // Switch (cond)
        {"vl", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            for _, c := range to_l(b) {
                truth := false
                var conseq interface{}
                if ! is_l(c) {
                    truth = true
                    conseq = c
                } else {
                    cl := to_l(c)
                    _case := cl[0]
                    switch {
                    case len(cl) == 1:
                        truth = true
                        cl = append(wrapa(0), cl...)    // Insert a placeholder so that the
                                                        // next step can correctly process it
                    case is_b(_case):
                        env.Push(a)
                        to_b(_case).Run(env)
                        truth = to_bool(env.Pop())
                    case is_l(_case):
                        truth = find(to_l(_case), wrapa(a)) > -1
                    default:
                        truth = equals(a, _case)
                    }
                    conseq = cl[1]
                }
                if truth {
                    if is_b(conseq) {
                        to_b(conseq).Run(env)
                    } else {
                        env.Push(conseq)
                    }
                    return nil
                }
            }
            return nil
        }},

        // Ternary if
        {"vaa", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := env.Pop3()
            var conseq interface{}
            if to_bool(a) {
                conseq = b
            } else {
                conseq = c
            }
            if is_b(conseq) {
                to_b(conseq).Run(env)
            } else {
                x.Push(conseq)
            }
            return nil
        }},
    }})// >>>>
    addOp(&Op{"@",// <<<<
    []TypedFunc{
        // Rotate
        {"aaa", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            return wraps(a, c, b)
        }},
    }})// >>>>
    addOp(&Op{"[",// <<<<
    []TypedFunc{
        // Start list
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            env.Mark()
            return wraps()
        }}}})
// >>>>
    addOp(&Op{"\\",// <<<<
    []TypedFunc{
        // Swap
        {"aa", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(a, b)
        }},
    }})// >>>>
    addOp(&Op{"]",// <<<<
    []TypedFunc{
        // End list
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            env.PopMark()
            return wraps()
        }}}})
// >>>>
    addOp(&Op{"^",// <<<<
    []TypedFunc{
        // Bitwise XOR
        {"ii", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) ^ to_i(b))
        }},

        {"cc", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) ^ to_i(b))
        }},

        {"ic", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(to_i(a) ^ to_i(b)))
        }},

        // Symmetric set diff
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bl := wrapa(), wrapa()
            res := wrapa()
            if is_l(a) {
                al = to_l(a)
            } else {
                al = wrapa(a)
            }
            if is_l(b) {
                bl = to_l(b)
            } else {
                bl = wrapa(b)
            }
            for _, i := range al {
                if find(bl, wrapa(i)) == -1 && find(res, wrapa(i)) == -1 {
                    res = append(res, i)
                }
            }
            for _, i := range bl {
                if find(al, wrapa(i)) == -1 && find(res, wrapa(i)) == -1 {
                    res = append(res, i)
                }
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"_",// <<<<
    []TypedFunc{
        // Duplicate
        {"a", 0x00,
        func(env * Environ, x * Stack) *Stack {
            x.Push(x.Get(0))
            return nil
        }},
    }})
// >>>>
    addOp(&Op{"|",// <<<<
    []TypedFunc{
        // Bitwise OR
        {"ii", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) | to_i(b))
        }},

        {"ic", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(to_i(a) | to_i(b)))
        }},

        // Set union
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bl := wrapa(), wrapa()
            res := wrapa()
            if is_l(a) {
                al = to_l(a)
            } else {
                al = wrapa(a)
            }
            if is_l(b) {
                bl = to_l(b)
            } else {
                bl = wrapa(b)
            }
            for _, i := range al {
                if find(res, wrapa(i)) == -1 {
                    res = append(res, i)
                }
            }
            for _, i := range bl {
                if find(res, wrapa(i)) == -1 {
                    res = append(res, i)
                }
            }
            return wraps(res)
        }},

        // If-else
        {"vb", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := env.Pop2()
            if ! to_bool(a) {
                to_b(b).Run(env)
            }
            return nil
        }},
    }})// >>>>
    addOp(&Op{"~",// <<<<
    []TypedFunc{
        // Bitwise NOT
        {"i", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(^to_i(x.Get1()))
        }},

        // Evaluate block/string/char, or dump list
        {"a", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a := env.Pop()
            switch {
            case is_b(a):
                to_b(a).Run(env)
            case is_s(a) || is_c(a):
                parse(NewParser("<string>", to_s(a)), false).Run(env)
            case is_l(a):
                unwrapped := NewStack(to_l(a))
                unwrapped.Reverse()
                env.Pusha(unwrapped)
            }
            return nil
        }},
    }})// >>>>
// >>>>
    // Alphabet operators// <<<<
    addOp(&Op{"a",// <<<<
    []TypedFunc{
        // Wrap in list
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(wrapa(x.Get1()))
        }},
    }})// >>>>
    addOp(&Op{"b",// <<<<
    []TypedFunc{
        // Convert to base
        {"ii", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ai, bi := to_i(a), to_i(b)
            res := wrapa()
            for ai > 0 {
                res = append(wrapa(ai % bi), res...)
                ai /= bi
            }
            return wraps(res)
        }},

        // Convert from base
        {"li", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            res := 0
            for _, i := range al {
                res = res * bi + to_i(i)
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"c",// <<<<
    []TypedFunc{
        // Convert to char
        {"p", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_c(x.Get1()))
        }},

        // Get first char
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_c(rune(to_s(x.Get1())[0])))     // Was byte, so must cast to rune
        }},
    }})// >>>>
    addOp(&Op{"d",// <<<<
    []TypedFunc{
        // Convert to double
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_d(x.Get1()))
        }},
    }})// >>>>
    // TODO `e` The extended operators (misc module)
    addOp(&Op{"g",// <<<<
    []TypedFunc{
        // Do-while (popping the condition)
        {"b", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a := env.Pop()
            ab := to_b(a)
            ab.Run(env)
            for to_bool(env.Pop()) {
                ab.Run(env)
            }
            return nil
        }},

        // TODO Get from URL (See test.go)
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            url := to_s(x.Get1())
            if ! strings.Contains(url, "://") {
                url = "http://" + url
            }
            resp, _ := http.Get(url)
            s, _ := ioutil.ReadAll(resp.Body)
            resp.Body.Close()
            return wraps(string(s))
        }},
    }})// >>>>
    addOp(&Op{"h",// <<<<
    []TypedFunc{
        // Do-while (without popping the condition)
        {"b", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a := env.Pop()
            ab := to_b(a)
            ab.Run(env)
            for to_bool(x.Get(0)) {
                ab.Run(env)
            }
            return nil
        }},
    }})// >>>>
    addOp(&Op{"i",// <<<<
    []TypedFunc{
        // Convert to int
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(x.Get1()))
        }},
    }})// >>>>
    addOp(&Op{"l",// <<<<
    []TypedFunc{
        // Readline
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            input := bufio.NewScanner(os.Stdin)
            if input.Scan() {
                return wraps([]rune(input.Text()))
            }
            return wraps("\n")
        }},
    }})// >>>>
    // TODO `m` The Math module
    // The 'os' module// <<<<
    // Brief:// <<<<
    // Use file objects???
    // oa: Append to file
    // oa/oA: CLI Arguments (maybe in extend operators?)
    // oc/oC: Change directory
    // od: create Directory
    // oe: file Exists?
    // of: create File
    // og/or: read from file (Get)
    // oh/op/oW: working directory (consider as Home?)
    // ol: List directory
    // om: Move file/directory
    // on/oo: chowN
    // oo/oM: chmOd
    // op/oc: coPy file/directory
    // oq: exit (Quit)
    // or/oR: Remove file/directory
    // os: Symlink
    // ot: current Time
    // ou: Unix time
    // ow: Write to file
    // ox: eXecute
    // oX: eXecute with specified input and output as string// >>>>
    addOp(&Op{"oa",// <<<<
    []TypedFunc{
        // Append to file
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            f, err := os.OpenFile(to_s(a), os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0666)
            defer f.Close()
            if err != nil {
                panic(err.Error())
            }
            bs := to_s(b)
            _, err = f.WriteString(bs)
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"oc",// <<<<
    []TypedFunc{
        // Change directory
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            err := os.Chdir(to_s(a))
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"od",// <<<<
    []TypedFunc{
        // Create directory (mkdir -p)
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            err := os.MkdirAll(to_s(x.Get1()), 0775)
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"oe",// <<<<
    []TypedFunc{
        // Check if file exists (test -f)
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            _, err := os.Stat(to_s(x.Get1()))
            return wraps(to_i(!os.IsNotExist(err)))
        }},
    }})// >>>>
    addOp(&Op{"of",// <<<<
    []TypedFunc{
        // Create file
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            f, err := os.Create(to_s(x.Get1()))
            defer f.Close()
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"og",// <<<<
    []TypedFunc{
        // Read file
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            d, err := ioutil.ReadFile(to_s(x.Get1()))
            if err != nil {
                panic(err.Error())
            }
            return wraps(string(d))
        }},
    }})// >>>>
    addOp(&Op{"oh",// <<<<
    []TypedFunc{
        // Get working directory (pwd)
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            dir, err := os.Getwd()
            if err != nil {
                panic(err.Error())
            }
            return wraps(dir)
        }},
    }})// >>>>
    addOp(&Op{"ol",// <<<<
    []TypedFunc{
        // List directory
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            files, err := ioutil.ReadDir(to_s(x.Get1()))
            if err != nil {
                panic(err.Error())
            }
            res := wrapa()
            for _, f := range files {
                res = append(res, f.Name())
            }
            return wraps(res)
        }},

        {"c", 0x10,
        func(env * Environ, x * Stack) *Stack {
            files, err := ioutil.ReadDir(to_s(x.Get1()))
            if err != nil {
                panic(err.Error())
            }
            res := wrapa()
            for _, f := range files {
                res = append(res, f.Name())
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"om",// <<<<
    []TypedFunc{
        // Move file to destination (mv)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            exe("mv", to_s(a), to_s(b))
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"on",// <<<<
    []TypedFunc{
        // Change owner (chown)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            usr, err := user.Lookup(to_s(b))
            if err != nil {
                panic(err.Error())
            }
            uid, _ := strconv.Atoi(usr.Uid)
            gid, _ := strconv.Atoi(usr.Gid)
            err = os.Chown(to_s(a), uid, gid)
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"oo",// <<<<
    []TypedFunc{
        // Change mode of file (chmod)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            exe("chmod", to_s(b), to_s(a))
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"op",// <<<<
    []TypedFunc{
        // Copy file to destination (cp)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            exe("cp", to_s(a), to_s(b))
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"oq",// <<<<
    []TypedFunc{
        // Exit (exit)
        {"i", 0x10,
        func(env * Environ, x * Stack) *Stack {
            os.Exit(to_i(x.Get1()))
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"or",// <<<<
    []TypedFunc{
        // Remove file or directory (rm)
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            exe("rm", "-rf", to_s(x.Get1()))
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"os",// <<<<
    []TypedFunc{
        // Create symlink (ln -s)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            err := os.Symlink(to_s(a), to_s(b))
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"ot",// <<<<
    []TypedFunc{
        // Time
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            now := time.Now()
            yr, mo, day := now.Date()
            hr, min, sec := now.Clock()
            nano := now.Nanosecond() / 1000000  // Get milliseconds
            _, zone := now.Zone()
            zone /= 3600        // Get hours
            return wraps(wrapa(yr, int(mo), day, hr, min, sec, nano, zone))
        }},
    }})// >>>>
    addOp(&Op{"ou",// <<<<
    []TypedFunc{
        // Milliseconds since Epoch (date +%s)
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(int(time.Now().UnixNano() / 1000000))
        }},
    }})// >>>>
    addOp(&Op{"ow",// <<<<
    []TypedFunc{
        // Write to file
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            f, err := os.OpenFile(to_s(a), os.O_WRONLY | os.O_CREATE, 0666)
            defer f.Close()
            if err != nil {
                panic(err.Error())
            }
            bs := to_s(b)
            _, err = f.WriteString(bs)
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"ox",// <<<<
    []TypedFunc{
        // Execute command with args
        {"sl", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            args := make([]string, 0)
            for _, i := range to_l(b) {
                args = append(args, to_s(i))
            }
            cmd := exec.Command(to_s(a), args...)
            output, err := cmd.CombinedOutput()
            fmt.Print(string(output))
            if err != nil {
                panic(err.Error())
            }
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"oX",// <<<<
    []TypedFunc{
        {"sls", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            args := make([]string, 0)
            for _, i := range to_l(b) {
                args = append(args, to_s(i))
            }
            cmd := exec.Command(to_s(a), args...)
            stdin, err := cmd.StdinPipe()
            defer stdin.Close()
            if err != nil {
                panic(err.Error())
            }

            _, err = io.WriteString(stdin, to_s(c))
            if err != nil {
                panic(err.Error())
            }
            output, err := cmd.CombinedOutput()
            if err != nil {
                panic(err.Error())
            }
            return wraps(string(output))
        }},
    }})//>>>>//>>>>
    addOp(&Op{"p",// <<<<
    []TypedFunc{
        // Print (no newline)
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            fmt.Print(to_s(x.Get1()))
            return wraps()
        }},
    }})// >>>>
    addOp(&Op{"q",// <<<<
    []TypedFunc{
        // Read all input
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            data, _ := ioutil.ReadAll(os.Stdin)
            return wraps([]rune(string(data)))
        }},
    }})// >>>>
    // TODO `r` The `regex` module// <<<<
    // rm: Checks if s matchs the pattern /^pat$/
    // rf: Trys to find a match matching the pattern /pat/, and return true if finds
    // rs: Like rf, but return a string
    // rl: Like rf, but return a list containing the match if exists, and the start and end
    //     positions
    // ra: Like rs, but find all matches, as if with "g" flag
    // re: return everything, like `ra` and `rl` combined// >>>>
    addOp(&Op{"s",// <<<<
    []TypedFunc{
        // Stringify
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_s(x.Get1()))
        }},
    }})// >>>>
    addOp(&Op{"t",// <<<<
    []TypedFunc{
        // Set list item
        {"lia", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            al := to_l(a)
            al[adjust_indm(to_i(b), len(al))] = c
            return wraps(al)
        }},
    }})// >>>>
    addOp(&Op{"v",// <<<<
    []TypedFunc{
        // Value representation
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_v(x.Get1()))
        }},
    }})// >>>>
    addOp(&Op{"w",// <<<<
    []TypedFunc{
        // While
        {"bb", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ab, bb := to_b(a), to_b(b)
            for {
                ab.Run(env)
                if ! to_bool(env.Pop()) {
                    break
                }
                bb.Run(env)
            }
            return wraps()
        }},
    }})// >>>>
    // TODO `x` The `Xfer` module// <<<<
    // xp: HTTP POST
    // xg: HTTP GET
    // xj: JSON encode
    // xk: JSON decode (maybe xJ?)
    // xx: XML encode
    // xy: XML decode (maybe xX?)// >>>>
    addOp(&Op{"y",// <<<<
    []TypedFunc{
        // Recurse function (No signature because it's not fixed)
        {"", 0x00,
        func(env * Environ, x * Stack) *Stack {
            defer func() {
                // Handle the index err paniced by the `Get` and `Pop` functions
                if err := recover(); err != nil {
                    if strings.Contains(fmt.Sprint(err), "bounds") {
                        panic("Not enough arguments to call `y`!")
                    } else {
                        panic(err)
                    }
                }
            }()

            if m := env.GetMemo(); m != nil {
                m.Run(env)
                return nil
            }
            arity := 1
            var block *Block
            a := env.Pop()
            if is_i(a) {
                arity = to_i(a)
                block = to_b(env.Pop())
            } else if is_b(a) {
                block = to_b(a)
            } else {
                panic(fmt.Sprintf("%T `y` not implemented!", typeof(a)))
            }
            //cache := env.Pop()
            memo := NewMemo(block, arity)
            env.SetMemo(memo)
            memo.Run(env)
            env.SetMemo(nil)
            return nil
        }},
    }})// >>>>
    addOp(&Op{"z",// <<<<
    []TypedFunc{
        // Absolute value
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            ad := to_d(a)
            res := math.Abs(ad)
            if is_i(a) {
                return wraps(to_i(res))
            }
            return wraps(to_d(res))
        }},

        // Transpose
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            al := to_l(a)
            res := [][]interface{}{}
            max_len := 0
            for _, l := range al {
                length := len(to_l(l))
                if length > max_len {
                    max_len = length
                }
            }

            for j := 0; j < max_len; j ++ {
                res_row := wrapa()
                for _, l := range al {
                    ll := to_l(l)
                    if len(ll) > j {
                        res_row = append(res_row, ll[j])
                    }
                }
                res = append(res, res_row)
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"xx",// <<<<
    []TypedFunc{
        {"a", 0x00,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get(0)
            fmt.Printf("\x1b[33mDEBUG\x1b[0m type: %s\n", fulltype(a))
            return nil
        }},
    }})// >>>>
// >>>>
}
//>>>>
// Helper functions// <<<<
func preproc(l []interface{}) (p []int) {
    subl := len(l)
    p = append(p, -1)
    for i, n := 1, 0; i < subl; i, n = i + 1, n + 1 {
        if equals(l[i], l[n]) {
            p = append(p, p[n])
        } else {
            p = append(p, n)
            n = p[n]
            for n >= 0 && ! equals(l[i], l[n]) {
                n = p[n]
            }
        }
    }
    return
}

func equals(a, b interface{}) bool {
    if ! (typeof(a) == typeof(b)) {
        return false
    }

    if is_l(a) {
        al, bl := to_l(a), to_l(b)
        if len(al) != len(bl) {
            return false
        }
        for i := 0; i < len(al); i ++ {
            if ! equals(al[i], bl[i]) {
                return false
            }
        }
        return true
    }

    return a == b
}

func split(s, sub []interface{}, empty bool) []interface{} {
    p := preproc(sub)
    sl := len(s)
    max := len(sub) - 1

    l := make([]interface{}, 0)
    x := 0
    for i, m := 0, 0; i < sl; i, m = i + 1, m + 1 {
        for m >= 0 && ! equals(sub[m], s[i]) {
            m = p[m]
        }
        if m == max {
            t := s[x:i-m]
            if empty || ! (len(t) == 0) {
                l = append(l, t)
            }
            x = i + 1
            m = -1
        }
    }
    t := s[x:len(s)]
    if empty || ! (len(t) == 0) {
        l = append(l, t)
    }
    return l
}

func find(s, sub []interface{}) int {
    p := preproc(sub)
    sl := len(s)
    max := len(sub) - 1

    for i, m := 0, 0; i < sl; i, m = i + 1, m + 1 {
        for m >= 0 && ! equals(sub[m], s[i]) {
            m = p[m]
        }
        if m == max {
            return i - m
        }
    }
    return -1
}

type Sorter struct {
    arr     []interface{}
    key     func(a interface{}) interface{}
}

func I(a interface{}) interface{} {     // Sorting function with no key
    return a
}

func NewSorter(arr []interface{}, key func(a interface{}) interface{}) *Sorter {
    return &Sorter{arr, key}
}

func (s * Sorter) Len() int {
    return len(s.arr)
}

func (s * Sorter) Less(i, j int) bool {
    return comp(s.key(s.arr[i]), s.key(s.arr[j])) < 0
}

func (s * Sorter) Swap(i, j int) {
    s.arr[i], s.arr[j] = s.arr[j], s.arr[i]
}

func comp(a, b interface{}) int {
    if is_l(a) && is_l(b) {
        al, bl := to_l(a), to_l(b)
        as, bs := len(al), len(bl)
        size := 0
        if as < bs {
            size = as
        } else {
            size = bs
        }
        for i := 0; i < size; i ++ {
            res := comp(al[i], bl[i])
            if res != 0 {
                return res
            }
        }
        switch {
        case as < bs:
            return -1
        case as > bs:
            return 1
        }
        return 0
    } else if (is_c(a) || is_n(a)) && (is_c(b) || is_n(b)) {
        res := to_d(a) - to_d(b)
        switch {
        case res < 0:
            return -1
        case res > 0:
            return 1
        default:
            return 0
        }
    }
    panic(fmt.Sprintf("Uncomparable types!: %T and %T", a, b))
}

func adjust_ind(ind, size int) int {
    if ind < 0 {
        ind += size
        if ind < 0 {
            ind = 0
        }
    } else if ind > size {
        ind = size
    }
    return ind
}

func adjust_indm(ind, size int) int {
    ind %= size
    fmt.Println("adjust_indm: ind % size:", ind)
    if ind < 0 {
        ind += size
    }
    return ind
}

func is_varchar(c rune) bool {
    return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_'
}

func is_upper(c rune) bool {
    return c >= 'A' && c <= 'Z'
}

func is_digit(c rune) bool {
    return c >= '0' && c <= '9'
}

func exe(cmd string, args ...string) {
    err := exec.Command(cmd, args...).Run()
    if err != nil {
        panic(err.Error())
    }
}
// >>>>
// Main// <<<<
func main() {
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, syscall.SIGINT)
    go func() {
        select {
        case <-signals:
            exit()
        }
    }()
    env := NewEnviron()
    InitFuncs()

    input := bufio.NewScanner(os.Stdin)
    fmt.Println("Cgam v1.0.0 by Steven.")
    fmt.Print(">>> ")
    for input.Scan() {
        var block *Block
        var parser *Parser
        code_str := input.Text()
        if code_str == "" {
            goto next
        }

        //// For debug
        //parser = NewParser("<stdin>", code_str)
        //block = parse(parser, false)
        //fmt.Println("Block:", block)
        //block.Run(env)
        //env.stack.Dump()
        //env.stack.Clear()

        // Parsing
        parser = NewParser("<stdin>", code_str)
        func() {
            defer func() {
                if err := recover(); err != nil {
                    fmt.Println("\x1b[31mERROR\x1b[0m: Parser:", err)
                    fmt.Printf(
`  at %s line %d:
    %s
    %s^
`,
                    parser.GetSrc(), parser.LnNum + 1,
                    strings.Split(string(parser.content), "\n")[parser.LnNum],
                    strings.Repeat(" ", parser.Offset - 1))
                }
            }()
            block = parse(parser, false)
        }()
        if block == nil {
            goto next
        }
        fmt.Println("Block:", block)

        // Running
        func() {
            defer func() {
                recover()   // Nothing, just empty `catch`
            }()
            block.Run(env)
            env.stack.Dump()
        }()
        env.stack.Clear()
next:
        fmt.Print(">>> ")
    }
    exit()
}

func exit() {
    fmt.Println("\nHappy jaming! :)")
    os.Exit(0)
}
// >>>>
// vim: set foldmethod=marker foldmarker=<<<<,>>>>
