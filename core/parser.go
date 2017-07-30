package cgam

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
)

// Parsing//
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

func Parse(code *Parser, withbrace bool) *Block {
    defer func() {
        if err := recover(); err != nil {
            fmt.Println("\x1b[31mERROR\x1b[0m: Parser:", err)
            fmt.Printf(
`  at %s line %d:
    %s
    %s^
`,
            code.GetSrc(), code.LnNum + 1,
            strings.Split(string(code.content), "\n")[code.LnNum],
            strings.Repeat(" ", code.Offset - 1))
        }
    }()
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
        return op_push(Parse(code, true))
    case '"':       // String
        str := ""
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
            str += string(char)
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
            return op_vector(Parse(code, true))
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
            panic("Invalid operator after `:`: " + op.String())
        }

    case '.':
        next, _ := code.Read()
        if next == '{' {
            return op_fold(Parse(code, true))
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
        panic("Invalid operator after `.`: " + op.String())

    case 'f':
        next, _ := code.Read()
        if next == '{' {
            return op_map2(Parse(code, true))
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
//
