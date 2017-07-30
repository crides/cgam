package main

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/Irides-Chromium/cgam/core"
)

// Main
const VERSION = "1.0"

const (
    CODE_FILE       = iota + 1
    CODE_IMMEDIATE
    //CODE_REPL
)

type Codetype struct {
    typ     int
    data    string
}

func NewCodetype(t int, d string) *Codetype {
    return &Codetype{t, d}
}

func usage() {
    fmt.Printf(
`Usage: %s [OPTION] [-- cgam_program_options]
  or %s
Run Cgam programs.

Options:
  -c, --code code_pieces    Treat the arguments as Cgam programs and run them.
  -f, --file files          Read programs from the files and run them.
  -i, --repl                Enter the REPL (Read-Eval-Print-Loop).
  -r, --reset RESET_OPTS    Reset the things in RESET_OPTS between executing different code pieces and between different lines in the REPL.

  -h, --help                Show this help message and exit.
  -v, --version             Show the version number and exit.

Note: When no option is provided, REPL would be entered by default. If '-c' or '-f' is used, the code pieces are executed in order. So if you want to use REPL after executing code from '-c' or '-f', you should specify a '-i' in the end and before '--', if it exists.
Options after '--' are passed to the underlying Cgam programs and can be accessed by using 'ea'.

RESET_OPTS can be one or more of stack, vars, lvars, ns, seperated by comma. If any is specified, then after executing a piece of code, the corresponding thing is reseted (or cleared). If nothing is to be reseted, then an empty string should be specified. Defaults to 'stack,vars'.
`, os.Args[0], os.Args[0])
    os.Exit(1)
}

func main() {
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, syscall.SIGINT)
    go func() {
        select {
        case <-signals:
            exit()
        }
    }()

    codes := make([]*Codetype, 0)
    cur_opt := 0
    var cgam_args []string
    _repl := true
    reset_opts := cgam.RESET_STACK | cgam.RESET_VARS
    // TODO Parse options
    for i := 1; i < len(os.Args); i ++ {
        opt := os.Args[i]
        switch opt {
        case "-v", "--version":
            fmt.Println("Cgam v" + VERSION)
            return
        case "-h", "--help":
            usage()
        case "-f", "--file":
            cur_opt = CODE_FILE
            _repl = false
        case "-c", "--code":
            cur_opt = CODE_IMMEDIATE
            _repl = false
        case "-i", "--repl":
            _repl = true
            cur_opt = 0
        case "-r", "--reset":
            r_opts := strings.Split(os.Args[i + 1], ",")
            i ++
            if len(r_opts) != 0 {
                reset_opts = 0
                for _, r_opt := range r_opts {
                    switch r_opt {
                    case "stack":
                        reset_opts |= cgam.RESET_STACK
                    case "vars":
                        reset_opts |= cgam.RESET_VARS
                    case "lvars":
                        reset_opts |= cgam.RESET_LONGVARS
                    case "ns":
                        reset_opts |= cgam.RESET_NAMESPACE
                    default:
                        fmt.Println("Invalid reset option!")
                        return
                    }
                }
            }
        case "--":
            cgam_args = os.Args[i + 1:]     // Add 2 because we start from os.Args[1]
            goto end_opts
        default:
            if cur_opt == 0 {
                fmt.Printf("Invalid option %s!\n", opt)
                usage()
            } else {
                codes = append(codes, NewCodetype(cur_opt, opt))
            }
        }
    }

end_opts:
    env := cgam.NewEnviron(cgam_args)
    cgam.InitFuncs()
    icode_num := 1
    for _, code := range codes {
        switch code.typ {
        case CODE_FILE:
            exe_file(env, code.data, reset_opts)
        case CODE_IMMEDIATE:
            exe_code(env, code.data, icode_num, reset_opts)
            icode_num ++
        default:
            panic("Invalid option!")
        }
    }
    if _repl {
        repl(env, reset_opts)
    }
}

func exe_code(env * cgam.Environ, code string, code_num, r_opt int) {
    parser := cgam.NewParser(fmt.Sprintf("<Code#%d>", code_num), code)
    block := cgam.Parse(parser, false)
    if block != nil {
        func() {
            defer func() {
                recover()   // Nothing, just empty `catch`
            }()
            block.Run(env)
        }()
        env.Dump(cgam.DUMP_STRING)
        env.Clear(r_opt)
    }
}

func exe_file(env * cgam.Environ, fname string, r_opt int) {
    file_data, err := ioutil.ReadFile(fname)
    if err != nil {
        fmt.Println(err)
        return
    }
    parser := cgam.NewParser(fname, string(file_data))
    block := cgam.Parse(parser, false)
    if block != nil {
        func() {
            defer func() {
                recover()   // Nothing, just empty `catch`
            }()
            block.Run(env)
        }()
        env.Dump(cgam.DUMP_STRING)
        env.Clear(r_opt)
    }
}

func repl(env * cgam.Environ, r_opt int) {
    input := bufio.NewScanner(os.Stdin)
    fmt.Printf("Cgam v%s by Steven.\n", VERSION)
    fmt.Print(">>> ")
    for input.Scan() {
        var block *cgam.Block
        var parser *cgam.Parser
        code_str := input.Text()
        if code_str == "" {
            goto next
        }

        // Parsing
        parser = cgam.NewParser("<stdin>", code_str)
        block = cgam.Parse(parser, false)
        if block == nil {
            goto next
        }

        // Running
        func() {
            defer func() {
                recover()   // Nothing, just empty `catch`
            }()
            block.Run(env)
            env.Dump(cgam.DUMP_VERTICAL)
        }()
        env.Clear(r_opt)
next:
        fmt.Print(">>> ")
    }
    exit()
}

func exit() {
    fmt.Println("\nHappy jaming! :)")
    os.Exit(0)
}
