package main

import (
    "bufio"
    "fmt"
    "io"
    "math"
    "os"
    "os/signal"
    "reflect"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "syscall"
)
// TODO 
// 1. Implement signal catching, and pretty `Bye!`'s;
// The memo for recursive functions (`y`)// <<<<
type Memo struct {
    block   *Block
    arity   int
    cache   map[interface{}]interface{}
}

func NewMemo(init map[interface{}]interface{}, b *Block, n int) *Memo {
    return &Memo{b, n, init}
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
    return
}

func (x * Stack) Pop2() (a, b interface{}) {
    b, a = x.Pop(), x.Pop()
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

func (x * Stack) Cut(a, b int) *Stack {
    return NewStack((*x)[a:b])
}

func (x * Stack) Dump() {
    fmt.Println("--- Top of Stack ---")
    for _, i := range *x {
        s := to_s(i)
        fmt.Printf("%s | %v\n", s, i)
    }
    fmt.Println("--- Bottom of Stack ---")
}

func (x * Stack) Clear() {
    *x = []interface{}{}
}

func (x * Stack) Size() int {
    return len(*x)
}

func (x * Stack) Switch(a, b int) {
    (*x)[a], (*x)[b] = (*x)[b], (*x)[a]
}

func (x * Stack) Reverse() {    // Reverse in-place
    i, j := 0, x.Size() - 1
    for i <= j {
        x.Switch(i, j)
        i ++
        j --
    }
}
// >>>>
// The environment// <<<<
type Environ struct {
    stack   *Stack
    marks   []int
    memo    *Memo
    vars    []interface{}
    args    []string
}

func NewEnviron() *Environ {
    env := &Environ{NewStack(nil),
                    make([]int, 0),
                    nil,
                    make([]interface{}, 0),
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

func (env * Environ) Mark() {
    env.marks = append(env.marks, env.stack.Size())
}

func (env * Environ) PopMark() {
    mark := 0
    if len(env.marks) > 0 {
        mark, env.marks = env.marks[0], env.marks[1:]
    }
    length := env.stack.Size()
    end := length - mark
    wrapped := NewStack((*env.stack)[:end])
    wrapped.Reverse()       // For compat
    *env.stack = append(wrapa([]interface{}(*wrapped)), (*env.stack)[end:]...)
}

// Variables
func (env * Environ) getVar(c rune) interface{} {
    if c >= 'A' && c <= 'Z' {
        return env.vars[c - 'A']
    }
    panic("Invalid variable name")
}

func (env * Environ) InitVars() {
    for i := 0; i <= 10; i ++ {
        env.vars = append(env.vars, 10 + i)
    }
    env.vars = append(env.vars, wrapa([]rune{}, []rune{}, []rune{'\n'}, []rune{}, math.Pi, []rune{}, []rune{}, []rune{' '}, 0, 0, 0, -1, 1, 2, 3)...)
    //                                 L   M    N    O   P        Q   R    S   T  U  V  W   X  Y  Z
}
// >>>>
// Type predicates//<<<<
func type_of(a interface{}) string {
    switch {
    case is_l(a):
        if is_s(a) {
            return "string"
        }
        return "list"
    case is_c(a):
        return "char"
    case is_i(a):
        return "int"
    case is_d(a):
        return "double"
    case is_b(a):
        return "block"
    default:
        return "unknown"
    }
    //return reflect.TypeOf(a).String()
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
        panic(fmt.Sprintf("Invalid type!: %T", a))
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
    default:
        panic("Invalid typename!: `" + string(typ) + "`")
    }
}

func _type_eq(a interface{}, typ string) bool {
    return reflect.TypeOf(a).String() == typ
}

func both_of(a, b interface{}, typ string) bool {
    return _type_eq(a, typ) && _type_eq(b, typ)
}

func any_of(a, b interface{}, typ string) bool {
    return _type_eq(a, typ) || _type_eq(b, typ)
}

func is_n(a interface{}) bool {
    return is_i(a) || is_d(a)
}

func is_i(a interface{}) bool {
    return _type_eq(a, "int")
}

func is_d(a interface{}) bool {
    return _type_eq(a, "float64")
}

func is_c(a interface{}) bool {
    return _type_eq(a, "rune") || _type_eq(a, "int32")
}

func is_l(a interface{}) bool {
    return _type_eq(a, "[]interface {}") || _type_eq(a, "[]int32")
}

func is_s(a interface{}) bool {
    if _type_eq(a, "[]int32") {     //[]rune
        return true
    }

    if ! is_l(a) {
        return false
    }

    for _, i := range a.([]interface{}) {
        if ! is_c(i) {
            return false
        }
    }
    return true
}

func is_b(a interface{}) bool {
    return _type_eq(a, "*main.Block")
}

func both_double(a, b interface{}) bool {
    return both_of(a, b, "float64")
}

func any_double(a, b interface{}) bool {
    return any_of(a, b, "float64")
}
//>>>>
// Type conversions// <<<<
func not_convertible(a interface{}, typ string) {
    panic(fmt.Sprintf("Cannot convert %s (%T) to %s!", type_of(a), a, typ))
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
    default:
        return fmt.Sprint(b)
    }
}

func to_v(a interface{}) string {       // To value string
    switch b := a.(type) {
    case int, float64:
        return fmt.Sprint(b)
    case []rune:
        return fmt.Sprintf("%q", string(b))
    case rune:
        return "'" + string(b)
    case *Block:
        return b.String()
    default:
        if is_s(a) {
            return to_s(a)
        }
        panic(fmt.Sprintf("Invalid type %T!", b))
    }
}

func to_bool(a interface{}) bool {
    switch b := a.(type) {
    case int, rune:
        return ! (b == 0)
    case float64:
        return ! (b == 0)
    case []interface{}:
        return ! (len(b) == 0)
    case []rune:
        return ! (len(b) == 0)
    default:
        not_convertible(a, "bool")
    }
    return false
}
//>>>>
// Block//<<<<
type Block struct {
    Ops     []*Op
    Offsets [][2]int    //LineNum & Offset
}

func NewBlock(ops []*Op, offsets [][2]int) *Block {
    if len(ops) == len(offsets) {
        return &Block{ops, offsets}
    }
    panic("Unmatch sizes!")
}

func (b * Block) Run(env * Environ) {
    for _, op := range b.Ops {
        op.Call(env)
    }
}

func (b * Block) String() string {
    s := []byte{'{'}
    for _, op := range b.Ops {
        s = append(s, append([]byte(op.String()), ' ')...)
    }
    if s[len(s) - 1] == ' ' {
        s[len(s) - 1] = '}'
    } else {
        s = append(s, '}')
    }
    return string(s)
}
//>>>>
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
        []TypedFunc{{
            "",
            0x10,
            func(env * Environ, x * Stack) *Stack {
                return wraps(a)
            }}}}
}

func op_pushVar(c rune) *Op {
    return &Op{string(c),
        []TypedFunc{{
            "",
            0x00,
            func(env * Environ, x * Stack) *Stack {
                a := env.getVar(c)
                if ! is_b(a) {
                    x.Push(a)
                } else {
                    to_b(a).Run(env)
                }
                return nil
            }}}}
}

func match_sign(tf TypedFunc, x * Stack) int {
    // Return value:
    // 0: not match
    // 1: matched
    // 2: special match (match if switched)
    arity := len(tf.Sign)
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
        sw_s, sw_e := switch_opts & 0xC0, switch_opts & 0x03
        var new_tf TypedFunc
        new_sign[sw_s], new_sign[sw_e] = new_sign[sw_e], new_sign[sw_s]
        new_tf.Sign = string(new_sign)
        if match_sign(new_tf, x) > 0 {
            // A special match
            return 2
        }
    }
    // Not matched
    return 0
}

func (op * Op) Call(env * Environ) {
    // Call functions// <<<<
    // For simplicity, the top of the stack is args[0]
    // But the arguments passed to Op.Run() is in the reverse direction
    //           -----------
    // args[0]    Environ top     c
    //           -----------
    // args[1]       ...        b
    //           -----------
    // args[2]       ...        a
    //           -----------
    // args[...]     ...
    //           -----------
    // args[-1]    Bottom
// >>>>
    var func_no, func_arity byte = 0, 0     // For type compatibility down ...
    var match int       // The match flag;

    // Try to match signature
    for i, tf := range op.Funcs {
        match = match_sign(tf, env.stack)
        if match == 0 {      // Not matched
            goto next_func
        }

        // Signature matched
        func_no = byte(i)
        func_arity = byte(len(tf.Sign))
        goto end_func_choose
next_func:
    }
    // loop not end normally
    panic(fmt.Sprintf("%T %T `%s` not implemented!", env.stack.Get(0), env.stack.Get(1), op.Name))

end_func_choose:
    // Prepare the arguements; switch if needed
    op_func := op.Funcs[func_no]
    wrapped := (op_func.Option & 0x10) > 0
    switch_opts := op_func.Option & 0x0F
    sw_s, sw_e := switch_opts & 0xC0, switch_opts & 0x03

    var raw_args []interface{}
    var args *Stack
    if wrapped {
        for i := 0; i < int(func_arity); i ++ {
            raw_args = append(wrapa(env.stack.Pop()), raw_args...)
        }
        args = NewStack(raw_args)
        if match == 2 {      // Only do a switch if it's a special match
            args.Switch(int(sw_s), int(sw_e))
        }
    } else {
        args = env.stack
        if match == 2 {
            args.Switch(int(func_arity - sw_s - 1), int(func_arity - sw_e - 1))       // ... here
        }
    }

    // Calling the function
    result := op_func.Func(env, args)
    if wrapped {
        env.stack.Pusha(result)
    }
}
//>>>>
// Parsing//<<<<
var (
    PATT_DOUBLE, _ = regexp.Compile("-?\\d+(\\.\\d+)?")
    PATT_NUM, _ = regexp.Compile("-?\\d+")
)

type Parser struct {
    LineNumber  int
    Offset      int
    io.RuneScanner
}

func NewParser(reader io.RuneScanner) *Parser {
    return &Parser{0, 0, reader}
}

func (p * Parser) GetOffset() [2]int {
    return [2]int{p.LineNumber, p.Offset}
}

func (p * Parser) ReadRune() (c rune, e error) {
    c, _, e = p.RuneScanner.ReadRune()      // Call the native ReadRune
    if c == '\n' {
        p.LineNumber++
        p.Offset = 0
    }
    return
}

func (p * Parser) UnreadRune() error {
    p.Offset--
    return p.RuneScanner.UnreadRune()
}

func parse(code *Parser, withbrace bool) *Block {
    ops := make([]*Op, 0)
    offsets := make([][2]int, 0)

    for c, err := code.ReadRune(); ; c, err = code.ReadRune() {
        if err != nil {
            if withbrace {
                panic("Unfinished block")
            }
            return NewBlock(ops, offsets)
        }

        switch c {
        case ' ', '\t', '\n':
        case '}':
            if withbrace {
                return NewBlock(ops, offsets)
            }
            panic("Unexpected `}`")
        case ';':       // A line comment
            for c, err := code.ReadRune(); err == nil && c != '\n'; c, err = code.ReadRune() { }
        default:
            offsets = append(offsets, code.GetOffset())
            code.UnreadRune()
            op := parseOp(code)
            ops = append(ops, op)
        }
    }
}

func parseNumber(code *Parser, nega bool) *Op {
    var num_str []rune       // String repr for the result number
    float := false

    if nega {
        num_str = append(num_str, '-')
    }

    for c, err := code.ReadRune(); err == nil; c, err = code.ReadRune() {
        switch {
        case c >= '0' && c <= '9':
            num_str = append(num_str, c)
        case c == '.':
            if float {
                code.UnreadRune()
                goto end_parse
            }
            float = true
            num_str = append(num_str, c)
        default:
            code.UnreadRune()
            goto end_parse
        }
    }

end_parse:
    if float {
        num_f, _ := strconv.ParseFloat(string(num_str), 64)
        return op_push(num_f)
    }
    num_i, _ := strconv.ParseInt(string(num_str), 10, 0)
    return op_push(int(num_i))
}

func parseOp(code *Parser) *Op {
    char, err := code.ReadRune()
    if err != nil {
        panic("Expects operator!")
    }

    if char >= '0' && char <= '9' {
        code.UnreadRune()
        return parseNumber(code, false)
    }

    if char >= 'A' && char <= 'Z' {
        return op_pushVar(char)
    }

    switch char {
    case '{':       // Block
        return op_push(parse(code, true))
    case '"':       // String
        var str []rune
        for char, _ := code.ReadRune(); err == nil; char, _ = code.ReadRune() {
            if char == '"' {
                return op_push(str)
            }
            if char == '\\' {
                c, err := code.ReadRune()
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
        actual_char, err := code.ReadRune()
        if err != nil {
            panic("Unfinished char!")
        }
        return op_push(actual_char)

    case '-':
        if next, _ := code.ReadRune(); next >= '0' && next <= '9' {     // Check if next is digit
            code.UnreadRune()
            return parseNumber(code, true)
        }
        code.UnreadRune()       // Spit out the "possible digit"
        return findOp("-")

    default:
        return findOp(string(char))
    }
}
//>>>>
// Functions //<<<<
func InitFuncs() {
    // Note:
    // 1. The return value in the wrapped functions must be wraps()'ed;
    // 2. For non-wrapping functions, the `x` is actually the stack in `env`;
    // 3. For non-wrapping functions, the return value can be `nil`, because it is actually
    //    discarded;
    // 4. Sometimes a condition would be put in front of the others because checking the type
    //    first would substract that situation from another situation.
    // 5. The order of the operators is the order of the corespondent Unicode code-points.
    addOp(&Op{"!",// <<<<
    []TypedFunc{
        // Logical NOT
        {"a",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(! to_bool(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"#",// <<<<
    []TypedFunc{
        // Numeric power
        {"nn",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            res := math.Pow(to_d(a), to_d(b))
            if res == to_d(to_i(res)) && ! any_double(a, b) {
                return wraps(to_i(res))
            }
            return wraps(res)
        }},

        // Find index
        {"lv",
        0x11,
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

        // Find index that satisfy block TODO Test
        {"lb",
        0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Pop2()
            al := to_l(a)
            bb := to_b(b)
            for i := 0; i < len(al); i ++ {
                x.Push(al[i])
                bb.Run(env)
                if to_bool(x.Pop()) {
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
        {"p",
        0x00,
        func(env * Environ, x * Stack) *Stack {
            ind := to_i(x.Pop())
            if ind > 0 {
                x.Push(x.Get(ind))
            } else {
                x.Push(x.Get(x.Size() + ind))
            }
            return nil
        }},

        // Simple sort
        {"l",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            l := x.Get1()
            s := NewSorter(to_l(l), I)
            sort.Stable(s)
            return wraps(s.arr)
        }},

        // Sort by key TODO Test
        {"lb",
        0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Pop2()
            al, bb := to_l(a), to_b(b)

            s := NewSorter(al, func(a interface{}) interface{} {
                x.Push(a)
                bb.Run(env)
                return x.Pop()
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
        {"nn",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                //if to_d(a) < 0 {
                //    return wraps(-math.Remainder(-to_d(a), -to_d(b)))
                //}
                return wraps(math.Remainder(to_d(a), to_d(b)))
            }
            return wraps(to_i(a) % to_i(b))
        }},

        // Every nth item
        {"ln",
        0x11,
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
        {"lv",
        0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if is_l(b) {
                return wraps(split(to_l(a), to_l(b), false))
            } else {
                return wraps(split(to_l(a), wrapa(b), false))
            }
        }},

        // Foreach (wraps in a list)
        {"bv",
        0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Pop2()
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
        {"ii",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) & to_i(b))
        }},

        {"ic",
        0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(to_i(a) & to_i(b)))
        }},

        // Set intersection
        {"vv",
        0x10,
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
        {"vb",
        0x00,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Pop2()
            if to_bool(a) {
                to_b(b).Run(env)
            }
            return nil
        }},
    }})// >>>>
    addOp(&Op{"(",// <<<<
    []TypedFunc{
        // Decrement
        {"p",
        0x10,
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
        {"l",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            al := to_l(a)
            o, al := al[0], al[1:]
            return wraps(o, al)
        }},
    }})// >>>>
    addOp(&Op{")",// <<<<
    []TypedFunc{
        // Decrement
        {"p",
        0x10,
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

        // Uncons from left
        {"l",
        0x10,
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
        {"nn",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) * to_d(b))
            }
            return wraps(to_i(a) * to_i(b))
        }},

        // Repeat value
        {"vi",
        0x11,
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
        {"bi",
        0x01,
        func(env * Environ, x * Stack) *Stack {
            //b, a := x.Pop(), x.Pop()
            a, b := x.Pop2()
            for i := 0; int(i) < to_i(b); i ++ {
                to_b(a).Run(env)
            }
            return nil
        }},

        // Join
        {"lv",
        0x10,
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

        {"cl",
        0x10,
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
        {"lb",
        0x01,
        func(env * Environ, x * Stack) *Stack {
            b, a := x.Pop(), x.Pop()
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
        {"nn",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) + to_d(b))
            }
            return wraps(to_i(a) + to_i(b))
        }},

        // Character concatenation
        {"cc",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps([]rune{a.(rune), b.(rune)})
        }},

        // Character incrementation (-> Char)
        {"cn",
        0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(a) + to_c(b))
        }},

        // List concat
        {"ll",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(append(to_l(a), to_l(b)...))
        }},

        // List append
        {"la",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(append(to_l(a), b))
        }},

        // List append
        {"al",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(append(wrapa(a), to_l(b)...))
        }},
    }})
// >>>>
    addOp(&Op{"-",// <<<<
    []TypedFunc{
        // Numeric minus
        {"nn",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) - to_d(b))
            }
            return wraps(to_i(a) - to_i(b))
        }},

        // Character decrementation (-> Char)
        {"cn",
        0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_c(a) - to_c(b))
        }},

        // Character difference (-> int)
        {"cc",
        0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(to_i(a) - to_i(b))
        }},

        // Remove from list
        {"la",
        0x10,
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

        {"pl",
        0x10,
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
        {"nn",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                return wraps(to_d(a) / to_d(b))
            }
            return wraps(to_i(a) / to_i(b))
        }},

        // Split by length
        {"ln",
        0x11,
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
        {"lv",
        0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if is_l(b) {
                return wraps(split(to_l(a), to_l(b), true))
            } else {
                return wraps(split(to_l(a), wrapa(b), true))
            }
        }},

        // Foreach
        {"bv",
        0x01,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Pop2()
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
    addOp(&Op{"_",// <<<<
    []TypedFunc{
        {"",
        0x00,
        func(env * Environ, x * Stack) *Stack {
            x.Push(x.Get(0))
            return nil
        }},
    }})
// >>>>
    addOp(&Op{"[",// <<<<
    []TypedFunc{
        {"",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            env.Mark()
            return wraps()
        }}}})
// >>>>
    addOp(&Op{"]",// <<<<
    []TypedFunc{
        {"",
        0x10,
        func(env * Environ, x * Stack) *Stack {
            env.PopMark()
            return wraps()
        }}}})
// >>>>
    addOp(&Op{"t",// <<<<
    []TypedFunc{{
        "a",
        0x00,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get(0)
            fmt.Println("\x1b[33mDEBUG\x1b[0m type:", type_of(a))
            return nil
        }}}})
// >>>>
}
//>>>>
// Helper functions// <<<<
func preproc(l []interface{}) (p []int) {
    subl := len(l)
    p = append(p, -1)
    for i, n := 1, 0; i < subl; i ++ {
        if equals(l[i], l[n]) {
            p = append(p, p[n])
        } else {
            p = append(p, n)
            n = p[n]
            for n >= 0 && ! equals(l[i], l[n]) {
                n = p[n]
            }
        }
        n ++
    }
    return
}

func equals(a, b interface{}) bool {
    if ! (type_of(a) == type_of(b)) {
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
    for i, m := 0, 0; i < sl; i ++ {
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
        m ++
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

    for i, m := 0, 0; i < sl; i ++ {
        for m >= 0 && ! equals(sub[m], s[i]) {
            m = p[m]
        }
        if m == max {
            return i - m
        }
        m ++
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
        switch {
        case as < bs:
            return -1
        case as > bs:
            return 1
        }
        for i := 0; i < as; i ++ {
            res := comp(al[i], bl[i])
            if res != 0 {
                return res
            }
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

// >>>>
func main() {// <<<<
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

    var block *Block
    input := bufio.NewScanner(os.Stdin)
    fmt.Println("Cgam v1.0.0 by Steven.")
    fmt.Print(">>> ")
    for input.Scan() {
        code_str := input.Text()
        if code_str == "" {
            goto next
        }
        block = parse(NewParser(strings.NewReader(code_str)), false)
        fmt.Println("Block:", block)
        block.Run(env)
        env.stack.Dump()
        env.stack.Clear()
next:
        fmt.Print(">>> ")
    }
    exit()
}

func exit() {
    fmt.Println("\nHappy jaming! " + ":)")//string(0x1f601))
    os.Exit(0)
}
// >>>>
// vim: set foldmethod=marker foldmarker=<<<<,>>>>
