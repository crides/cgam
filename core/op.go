package cgam

import (
    "fmt"
    "math"
    "strings"
)

// Operator//
type TypedFunc struct {
    Sign    string        // Signature
    Option  byte          // * * * * | Switch_start Switch_end
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
            {"", 0x10,
            func(env * Environ, x * Stack) *Stack {
                return wraps(a)
            }},
        }}
}

func op_pushVar(s string) *Op {
    return &Op{":" + s,
    []TypedFunc{
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := env.GetVar(s)
            if ! is_b(a) {
                return wraps(a)
            }
            to_b(a).Run(env)
            return nil
        }},
    }}
}

func op_setVar(s string) *Op {
    return &Op{"." + s,
    []TypedFunc{
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            env.SetVar(s, env.Get(0))
            return nil
        }},
    }}
}

func op_vector(r Runner) *Op {
    return &Op{":" + r.String(),
    []TypedFunc{
        {"ll", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
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
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            env.Mark()
            for _, i := range to_l(x.Get1()) {
                env.Push(i)
                r.Run(env)
            }
            env.PopMark()
            return nil
        }},
    }}
}

func op_fold(r Runner) *Op {
    return &Op{"." + r.String(),
    []TypedFunc{
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            env.Push(al[0])
            for i := 1; i < len(al); i ++ {
                env.Push(al[i])
                r.Run(env)
            }
            return nil
        }},
    }}
}

func op_for(name string) *Op {
    return &Op{"f" + name,
    []TypedFunc{
        {"vb", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bb := wrapa(), to_b(b)
            if is_l(a) {
                al = to_l(a)
            } else {
                al = Range(to_i(a))
            }
            for _, i := range al {
                env.SetVar(name, i)
                bb.Run(env)
            }
            return nil
        }},
    }}
}

func op_map2(r Runner) *Op {
    return &Op{"f" + r.String(),
    []TypedFunc{
        {"la", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            for _, i := range to_l(a) {
                env.Push(i)
                env.Push(b)
                r.Run(env)
            }
            return nil
        }},

        {"vl", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
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
    // args[0]    Stack top     c
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
        sw_s, sw_e := int(op_func.Option & 0x0C), int(op_func.Option & 0x03)

        args := NewStack(nil)
        for i := 0; i < func_arity; i ++ {
            args.Push(env.Pop())
        }
        if match == 2 {      // Only do a switch if it's a special match
            args.Switch(sw_s, sw_e)
        }

        // Calling the function
        if result := op_func.Func(env, args); result != nil {
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
//
