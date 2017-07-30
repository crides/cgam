package cgam

import (
    "fmt"
    "strings"
)

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

const SIGNAL string = "!#\ufdd0"

func (b * Block) Run(env * Environ) {
    for i, op := range b.Ops {
        if i == 0 {         // Only setup handler once
            defer func() {
                if err := recover(); err != nil {
                    offsets := b.Offsets[i]
                    lnum := offsets[0]
                    column := offsets[1]
                    if err != SIGNAL {  // Error is a native error; prints header
                        fmt.Println("\x1b[31mERROR\x1b[0m: Runtime:", to_s(err))
                    }
                    fmt.Printf(
`  at %s line %d:
    %s
    %s^
`,
                    // Line number should start from 1.
                    b.parser.GetSrc(), lnum + 1,
                    strings.Split(string(b.parser.content), "\n")[lnum],
                    strings.Repeat(" ", column - 1))
                    panic(SIGNAL)   // Panics a dummy signal to upper levels to report error
                }
            }()
        }
        op.Run(env)
    }
}

func (b * Block) String() string {
    s := ""
    if b.withbrace {
        s += "{"
    }
    for _, op := range b.Ops {
        s += op.String() + " "
    }
    size := len(s)
    if size == 0 {
        return ""
    }
    if b.withbrace {
        if s[size - 1] == ' ' {
            s = s[:size - 1] + "}"
        } else {
            s += "}"
        }
    }
    return string(s)
}
//
type Runner interface {
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
}
