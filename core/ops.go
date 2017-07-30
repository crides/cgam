package cgam

import (
    "bufio"
    "encoding/json"
    "encoding/xml"
    "fmt"
    "io"
    "io/ioutil"
    "math"
    "net/http"
    "os"
    "os/exec"
    "os/user"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

// Functions
func InitFuncs() {
    // Note:
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
    // 5. The order of the operators is the order of the corespondent Unicode code-points.
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
            al, bl := to_l(a), to_l(b)
            return wraps(find(al, bl))
        }},

        // Find index that satisfy block
        {"lb", 0x11,
        func(env * Environ, x * Stack) *Stack {
            //a, b := x.Get2()
            a, b := x.Get2()
            al, bb := to_l(a), to_b(b)
            for i := 0; i < len(al); i ++ {
                env.Push(al[i])
                bb.Run(env)
                if to_bool(env.Pop()) {
                    return wraps(i)
                }
            }
            return wraps(-1)
        }},
    }})// >>>>
    addOp(&Op{"$",// <<<<
    []TypedFunc{
        // Copy from stack
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ind := to_i(x.Get1())
            if ind < 0 {
                return wraps(env.Get(env.Size() + ind))
            }
            return wraps(env.Get(ind))
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
        {"lb", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bb := to_l(a), to_b(b)

            s := NewSorter(al, func(a interface{}) interface{} {
                env.Push(a)
                bb.Run(env)
                return env.Pop()
            })
            sort.Stable(s)
            return wraps(s.arr)
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
            return wraps(split(to_l(a), to_l(b), false))
        }},

        // Foreach (wraps in a list)
        {"bv", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            env.Mark()
            ab, bl := to_b(a), wrapa()
            if is_c(b) || is_n(b) {
                bl = Range(to_i(b))
            } else {
                bl = to_l(b)
            }
            for _, i := range bl {
                env.Push(i)
                ab.Run(env)
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
            al, bl := to_l(a), to_l(b)
            res := wrapa()
            for _, i := range al {
                if find(bl, wrapa(i)) >= 0 && find(res, wrapa(i)) == -1 {
                    res = append(res, i)
                }
            }
            return wraps(res)
        }},

        // If-then
        {"vb", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
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
                s = append(s, to_l(a)...)
            }
            return wraps(s)
        }},

        // Repeat block execution
        {"bi", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
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
            bl := to_l(b)
            for i, o := range to_l(a) {
                if i > 0 {
                    res = append(res, to_l(bl)...)
                }
                res = append(res, to_l(o)...)
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
                res = append(res, to_l(o)...)
            }
            return wraps(res)
        }},

        // Fold / Reduce
        {"lb", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bb := to_l(a), to_b(b)
            if len(al) > 0 {
                env.Push(al[0])
                for i := 1; i < len(al); i ++ {
                    env.Push(al[i])
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
            return wraps(to_s(a) + to_s(b))
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
            return wraps(Range(n))
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
        {"lb", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bb := to_l(a), to_b(b)
            res := wrapa()
            for _, i := range al {
                env.Push(i)
                bb.Run(env)
                if to_bool(env.Pop()) {
                    res = append(res, i)
                }
            }
            return wraps(res)
        }},

        {"nb", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ai, bb := to_i(a), to_b(b)
            res := wrapa()
            for i := 0; i < ai; i ++ {
                env.Push(i)
                bb.Run(env)
                if to_bool(env.Pop()) {
                    res = append(res, i)
                }
            }
            return wraps(res)
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
            for _, i := range al {
                if find(to_l(b), wrapa(i)) == -1 {
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
            return wraps(split(to_l(a), to_l(b), true))
        }},

        // Foreach
        {"bv", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ab, bl := to_b(a), wrapa()
            if is_c(b) || is_n(b) {
                bl = Range(to_i(b))
            } else {
                bl = to_l(b)
            }
            for _, i := range bl {
                env.Push(i)
                ab.Run(env)
            }
            return nil
        }},
    }})
// >>>>
    addOp(&Op{"`",  // From `;` to ```// <<<<
    []TypedFunc{
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return nil
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
        {"lb", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bb := to_l(a), to_b(b)
            for _, i := range al {
                env.Push(i)
                bb.Run(env)
                if to_bool(env.Pop()) {
                    return wraps(i)
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
        {"vl", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            for _, c := range to_l(b) {
                truth := false
                var conseq interface{}
                if ! is_l(c) {
                    truth = true
                    conseq = c
                } else {
                    cl := to_l(c)
                    _cases := cl[:len(cl) - 1]
                    switch {
                    case len(_cases) == 0:
                        truth = true
                    case len(_cases) >= 1:
                        for _, _case := range _cases {
                            if is_b(_case) {
                                env.Push(a)
                                to_b(_case).Run(env)
                                truth = truth || to_bool(env.Pop())
                            } else {
                                truth = truth || equals(a, _case)
                            }
                        }
                    }
                    conseq = cl[len(cl) - 1]
                }
                if truth {
                    if ! is_b(conseq) {
                        return wraps(conseq)
                    }
                    to_b(conseq).Run(env)
                    return nil
                }
            }
            return nil
        }},

        // Ternary if
        {"vaa", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            var conseq interface{}
            if to_bool(a) {
                conseq = b
            } else {
                conseq = c
            }
            if ! is_b(conseq) {
                return wraps(conseq)
            }
            to_b(conseq).Run(env)
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
            return nil
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
            return nil
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
            al, bl := to_l(a), to_l(b)
            res := wrapa()
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
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(env.Get(0))
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
            al, bl := to_l(a), to_l(b)
            res := wrapa()
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
        {"vb", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
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
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            switch {
            case is_b(a):
                to_b(a).Run(env)
            case is_s(a) || is_c(a):
                Parse(NewParser("<string>", to_s(a)), false).Run(env)
            case is_l(a):
                unwrapped := NewStack(to_l(a))
                unwrapped.Reverse()
                return unwrapped
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
    // The extended operators (misc module)// <<<<
    addOp(&Op{"ea",// <<<<
    []TypedFunc{
        // CLI arguments
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            args := wrapa()
            for _, arg := range env.args {
                args = append(args, arg)
            }
            return wraps(args)
        }},
    }})// >>>>
    addOp(&Op{"ec",// From `e=` to `ec`; Added blocks // <<<<
    []TypedFunc{
        // Count occurences
        {"la", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al := to_l(a)
            count := 0
            if is_b(b) {
                for _, o := range al {
                    env.Push(o)
                    to_b(b).Run(env)
                    if to_bool(env.Pop()) {
                        count ++
                    }
                }
            } else {
                for _, o := range al {
                    if equals(o, b) {
                        count ++
                    }
                }
            }
            return wraps(count)
        }},
    }})// >>>>
    addOp(&Op{"ed",// <<<<
    []TypedFunc{
        // Debug (Dump stack)
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            env.Dump(DUMP_HORIZONTAL)
            return nil
        }},
    }})// >>>>
    addOp(&Op{"ee",// <<<<
    []TypedFunc{
        // Enumerate (like Python's)
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            res := wrapa()
            for i, o := range to_l(x.Get1()) {
                res = append(res, wrapa(i, o))
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"el",// <<<<
    []TypedFunc{
        // To lowercase
        {"c", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_c(strings.ToLower(to_s(x.Get1()))))
        }},

        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(strings.ToLower(to_s(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"et",// <<<<
    []TypedFunc{
        // Translate
        {"lll", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            al, bl, cl := to_l(a), to_l(b), to_l(c)
            cs := len(cl)
            res := wrapa()
            for _, o := range al {
                ind := find(bl, wrapa(o))
                switch {
                case ind < 0:
                    res = append(res, o)
                case ind < cs:
                    res = append(res, cl[ind])
                default:
                    res = append(res, cl[cs - 1])
                }
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"eu",// <<<<
    []TypedFunc{
        // To uppercase
        {"c", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_c(strings.ToUpper(to_s(x.Get1()))))
        }},

        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(strings.ToUpper(to_s(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"ew",// <<<<
    []TypedFunc{
        // Overlapping slices of list
        {"li", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            as := len(al)
            if bi <= 0 || bi > as {
                panic("Invalid size!")
            }
            res := wrapa()
            for i := 0; i < as - bi + 1; i ++ {
                res = append(res, al[i:i + bi])
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"e<", // From `m<` // <<<<
    []TypedFunc{
        // Left shift
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            bi := to_i(b)
            if bi < 0 {
                return findOp("e>").Funcs[0].Func(env, wraps(a, -bi))
            }
            if is_d(a) {
                return wraps(to_d(a) * to_d(1 << uint(bi)))
            }
            return wraps(to_i(a) << uint(bi))
        }},

        // Rotate left
        {"ln", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            as := len(al)
            if as < 2 {
                return wraps(al)
            }
            k := bi % as
            if k == 0 {
                return wraps(al)
            }
            return wraps(append(al[k:as], al[0:k]...))
        }},
    }})// >>>>
    addOp(&Op{"e>", // From `m>` // <<<<
    []TypedFunc{
        // Right shift
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            bi := to_i(b)
            if bi < 0 {
                return findOp("e<").Funcs[0].Func(env, wraps(a, -bi))
            }
            if is_d(a) {
                return wraps(to_d(a) / to_d(1 << uint(bi)))
            }
            return wraps(to_i(a) >> uint(bi))
        }},

        // Rotate right
        {"ln", 0x11,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            as := len(al)
            if as < 2 {
                return wraps(al)
            }
            k := bi % as
            if k == 0 {
                return wraps(al)
            }
            return wraps(append(al[as - k:as], al[0:as - k]...))
        }},
    }})// >>>>
    addOp(&Op{"e&",// <<<<
    []TypedFunc{
        // And
        {"va", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            var res interface{}
            if to_bool(a) {
                res = b
            } else {
                res = a
            }

            if is_b(res) {
                to_b(res).Run(env)
                return nil
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"e|",// <<<<
    []TypedFunc{
        // Or
        {"va", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            var res interface{}
            if to_bool(a) {
                res = a
            } else {
                res = b
            }

            if is_b(res) {
                to_b(res).Run(env)
                return nil
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"e%",// <<<<
    []TypedFunc{
        // Format
        {"sv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(fmt.Sprintf(to_s(a), to_l(b)...))
        }},
    }})// >>>>
    addOp(&Op{"e*",// <<<<
    []TypedFunc{
        // Repeat all
        {"ln", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            switch {
            case bi < 0:
                panic("Invalid size!")
            case bi == 0:
                return wraps(wrapa())
            case bi == 1:
                return wraps(al)
            default:
                res := wrapa()
                for _, o := range al {
                    for i := 0; i < bi; i ++ {
                        res = append(res, o)
                    }
                }
                return wraps(res)
            }
        }},
    }})// >>>>
    addOp(&Op{"e!",// <<<<
    []TypedFunc{
        // Permutations (no repeats)
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            s := NewSorter(al, I)
            sort.Stable(s)
            as := len(al)
            var k int
            res := wrapa(al)
            for {
                ol := al
                al = make([]interface{}, as)
                copy(al, ol)
                for k = as - 2; k >= 0; k-- {
                    if comp(al[k], al[k + 1]) < 0 {
                        break
                    }
                }
                if k < 0 {
                    break
                }
                i := as - 1
                for comp(al[k], al[i]) >= 0 {
                    i --
                }
                al[i], al[k] = al[k], al[i]
                k ++
                i = as - 1
                for k < i {
                    al[i], al[k] = al[k], al[i]
                    k ++
                    i --
                }
                res = append(res, al)
            }
            return wraps(res)
        }},

        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ai := to_i(x.Get1())
            l := wrapa()
            for i := 0; i < ai; i ++ {
                l = append(l, i)
            }
            return findOp("e!").Funcs[0].Func(env, wraps(l))
        }},
    }})// >>>>
    addOp(&Op{"e_",// <<<<
    []TypedFunc{
        // Flatten
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            return wraps(flatten(al))
        }},
    }})// >>>>
    addOp(&Op{"e`",// <<<<
    []TypedFunc{
        // Length-value encode
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            if len(al) == 0 {
                return wraps(al)
            }
            res := wrapa()
            o := al[0]
            n := 1
            for i := 1; i < len(al); i ++ {
                o1 := al[i]
                if equals(o, o1) {
                    n ++
                } else {
                    res = append(res, wrapa(n, o))
                    o = o1
                    n = 1
                }
            }
            res = append(res, wrapa(n, o))
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"e~",// <<<<
    []TypedFunc{
        // Length-value decode
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            if len(al) == 0 {
                return wraps(al)
            }
            res := wrapa()
            for _, o := range al {
                if ! is_l(o) {
                    panic("Need a list of pairs!")
                }
                ol := to_l(o)
                for i := 0; i < to_i(ol[0]); i ++ {
                    res = append(res, ol[1])
                }
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"e\\",// <<<<
    []TypedFunc{
        // Switch list items
        {"lii", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            al, bi, ci := to_l(a), to_i(b), to_i(c)
            as := len(al)
            ind1, ind2 := adjust_indm(bi, as), adjust_indm(ci, as)
            al[ind1], al[ind2] = al[ind2], al[ind1]
            return wraps(al)
        }},
    }})// >>>>
    addOp(&Op{"e[",// <<<<
    []TypedFunc{
        // Pad list left
        {"liv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            al, bi := to_l(a), to_i(b)
            as := len(al)
            if bi < 0 {
                panic("Invalid size!")
            }
            if bi <= as {
                return wraps(al)
            }
            res := al
            for i := 0; i < bi - as; i ++ {
                res = append(res, c)
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"e]",// <<<<
    []TypedFunc{
        // Pad list right
        {"liv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            al, bi := to_l(a), to_i(b)
            as := len(al)
            if bi < 0 {
                panic("Invalid size!")
            }
            if bi <= as {
                return wraps(al)
            }
            res := wrapa()
            for i := 0; i < bi - as; i ++ {
                res = append(res, c)
            }
            return wraps(append(res, al...))
        }},
    }})// >>>>// >>>>
    addOp(&Op{"g",// <<<<
    []TypedFunc{
        // Do-while (popping the condition)
        {"b", 0x10,
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
        {"b", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ab := to_b(x.Get1())
            ab.Run(env)
            for to_bool(env.Get(0)) {
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
                return wraps(input.Text())
            }
            return wraps("\n")
        }},
    }})// >>>>
    // The Math module// <<<<
    addOp(&Op{"m<",// From e< // <<<<
    []TypedFunc{
        // Min
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if comp(a, b) < 0 {
                return wraps(a)
            }
            return wraps(b)
        }},
    }})// >>>>
    addOp(&Op{"m>",// From e> // <<<<
    []TypedFunc{
        // Max
        {"vv", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if comp(a, b) > 0 {
                return wraps(a)
            }
            return wraps(b)
        }},
    }})// >>>>
    addOp(&Op{"mr",// <<<<
    []TypedFunc{
        // Random number
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            src := env.GetRand()
            if is_d(a) {
                return wraps(src.Float64() * to_d(a))
            }
            ai := to_i(a)
            if ai <= 0 {
                panic("Range must be positive!")
            }
            return wraps(src.Intn(ai))
        }},

        // Shuffle list
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            src := env.GetRand()
            for i := range al {
                j := src.Intn(i + 1)
                al[i], al[j] = al[j], al[i]
            }
            return wraps(al)
        }},
    }})// >>>>
    addOp(&Op{"mR",// <<<<
    []TypedFunc{
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            as := len(al)
            if as == 0 {
                panic("Empty list!")
            }
            return wraps(al[env.GetRand().Intn(as)])
        }},
    }})// >>>>
    addOp(&Op{"md",// <<<<
    []TypedFunc{
        // Modulus & division
        {"nn", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            if any_double(a, b) {
                ad, bd := to_d(a), to_d(b)
                return wraps(math.Remainder(ad, bd), ad / bd)
            }
            ai, bi := to_i(a), to_i(b)
            return wraps(ai % bi, ai / bi)
        }},
    }})// >>>>
    addOp(&Op{"m*",// <<<<
    []TypedFunc{
        // Cartesian product/power
        {"ll", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            return wraps(carte_product(to_l(a), to_l(b), true))
        }},

        {"li", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            al, bi := to_l(a), to_i(b)
            res := make([]interface{}, len(al))
            copy(res, al)
            for i := 1; i < bi; i ++ {
                res = carte_product(res, al, false)
            }
            return wraps(res)
        }},

        {"ii", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ai, bi := to_i(a), to_i(b)
            return findOp("m*").Funcs[1].Func(env, wraps(Range(ai), bi))
        }},
    }})// >>>>
    addOp(&Op{"mp",// <<<<
    []TypedFunc{
        // Prime?
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ai := to_i(x.Get1())
            switch {
            case ai == 2 || ai == 3:
                return wraps(1)
            case ai < 5 || ai % 2 == 0:
                return wraps(0)
            default:
                for i := 3; i * i <= ai; i += 2 {
                    if ai % i == 0 {
                        return wraps(0)
                    }
                }
                return wraps(1)
            }
        }},
    }})// >>>>
    addOp(&Op{"mf",// <<<<
    []TypedFunc{
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ai := to_i(x.Get1())
            res := wrapa()
            if ai < 4 {
                return wraps(wrapa(ai))
            }
            for ai % 2 == 0 {
                res = append(res, 2)
                ai >>= 1
            }
            for i := 3; i * i <= ai; i += 2 {
                for ai % i == 0 {
                    res = append(res, i)
                    ai /= i
                }
            }
            if ai > 1 {
                res = append(res, ai)
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"mF",// <<<<
    []TypedFunc{
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ai := to_i(x.Get1())
            res := wrapa()
            if ai < 4 {
                return wraps(wrapa(ai, 1))
            }

            n := 0
            for ai % 2 == 0 {
                n ++
                ai >>= 1
            }
            if n > 0 {
                res = append(res, wrapa(2, n))
            }

            for i := 3; i * i <= ai; i += 2 {
                n = 0
                for ai % i == 0 {
                    n ++
                    ai /= i
                }
                if n > 0 {
                    res = append(res, wrapa(i, n))
                }
            }
            if ai > 1 {
                res = append(res, wrapa(ai, 1))
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"mo",// <<<<
    []TypedFunc{
        // Round
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a := x.Get1()
            if is_d(a) {
                return wraps(round(to_d(a), 1))
            }
            return wraps(a)
        }},
    }})// >>>>
    addOp(&Op{"mO",// <<<<
    []TypedFunc{
        // Round to precision
        {"ni", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            res := round(to_d(a), math.Pow(10, to_d(b)))
            if res == to_d(to_i(res)) {
                return wraps(to_i(res))
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"ms",// <<<<
    []TypedFunc{
        // Sine
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Sin(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"mc",// <<<<
    []TypedFunc{
        // Cosine
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Cos(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"mt",// <<<<
    []TypedFunc{
        // Tangent
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Tan(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"mS",// <<<<
    []TypedFunc{
        // Arc-sine
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Asin(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"mC",// <<<<
    []TypedFunc{
        // Arc-cosine
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Acos(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"mT",// <<<<
    []TypedFunc{
        // Arc-tangent
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Atan(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"ma",// <<<<
    []TypedFunc{
        // Arc-tangent2
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ad, bd := to_d(a), to_d(b)
            return wraps(math.Atan2(ad, bd))
        }},
    }})// >>>>
    addOp(&Op{"mh",// <<<<
    []TypedFunc{
        // Hypot
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ad, bd := to_d(a), to_d(b)
            return wraps(math.Hypot(ad, bd))
        }},
    }})// >>>>
    addOp(&Op{"mq",// <<<<
    []TypedFunc{
        // Square root
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(math.Sqrt(to_d(x.Get1())))
        }},
    }})// >>>>
    addOp(&Op{"mQ",// <<<<
    []TypedFunc{
        // Integral square root
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(math.Sqrt(to_d(x.Get1()))))
        }},
    }})// >>>>
    addOp(&Op{"me",// <<<<
    []TypedFunc{
        // Base-e exponential
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(math.Exp(to_d(x.Get1()))))
        }},
    }})// >>>>
    addOp(&Op{"ml",// <<<<
    []TypedFunc{
        // Base-e logarithm
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(math.Log(to_d(x.Get1()))))
        }},
    }})// >>>>
    addOp(&Op{"mL",// <<<<
    []TypedFunc{
        // Logarithm (Log(a) b)
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            ad, bd := to_d(a), to_d(b)
            if ad == 10 {
                return wraps(math.Log10(bd))
            }
            return wraps(math.Log(bd) / math.Log(ad))
        }},
    }})// >>>>
    addOp(&Op{"m[",// <<<<
    []TypedFunc{
        // Floor
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(math.Floor(to_d(x.Get1()))))
        }},
    }})// >>>>
    addOp(&Op{"m]",// <<<<
    []TypedFunc{
        // Ceiling
        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            return wraps(to_i(math.Ceil(to_d(x.Get1()))))
        }},
    }})// >>>>
    addOp(&Op{"m!",// <<<<
    []TypedFunc{
        // Permutations (with repeats)
        {"l", 0x10,
        func(env * Environ, x * Stack) *Stack {
            al := to_l(x.Get1())
            as := len(al)
            p := findOp("e!").Funcs[1].Func(env, wraps(as))
            res := wrapa()
            for _, o := range to_l(p.Contents()[0]) {
                ol := to_l(o)
                for i := 0; i < as; i ++ {
                    ol[i] = al[to_i(ol[i])]
                }
                res = append(res, ol)
            }
            return wraps(res)
        }},

        {"n", 0x10,
        func(env * Environ, x * Stack) *Stack {
            ai := to_i(x.Get1())
            t := 1
            for i := 2; i < ai; i ++ {
                t *= i
            }
            return wraps(t)
        }},
    }})// >>>>
// >>>>
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
            return nil
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
            return nil
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
            return nil
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
            return nil
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
            return nil
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
            return nil
        }},
    }})// >>>>
    addOp(&Op{"oo",// <<<<
    []TypedFunc{
        // Change mode of file (chmod)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            exe("chmod", to_s(b), to_s(a))
            return nil
        }},
    }})// >>>>
    addOp(&Op{"op",// <<<<
    []TypedFunc{
        // Copy file to destination (cp)
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            exe("cp", to_s(a), to_s(b))
            return nil
        }},
    }})// >>>>
    addOp(&Op{"oq",// <<<<
    []TypedFunc{
        // Exit (exit)
        {"i", 0x10,
        func(env * Environ, x * Stack) *Stack {
            os.Exit(to_i(x.Get1()))
            return nil
        }},
    }})// >>>>
    addOp(&Op{"or",// <<<<
    []TypedFunc{
        // Remove file or directory (rm)
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            exe("rm", "-rf", to_s(x.Get1()))
            return nil
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
            return nil
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
            return nil
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
            return nil
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
            return nil
        }},
    }})// >>>>
    addOp(&Op{"q",// <<<<
    []TypedFunc{
        // Read all input
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            data, _ := ioutil.ReadAll(os.Stdin)
            return wraps(string(data))
        }},
    }})// >>>>
    // The `regex` module// <<<<
    // Brief:// <<<<
    // rm: Checks if s matchs the pattern /^pat$/
    // rf: Trys to find a match matching the pattern /pat/, and return true if finds
    // rr: Like rf, but return the result (a string)
    // rl: Like rf, but return a list containing the match if exists, and the start and end
    //     positions
    // ra: Like rr, but returns a list of all matches, as if with "g" flag
    // re: return everything, like `ra` and `rl` combined
    // ru/rA: Like rf, but with flags. `pat` `src` `flags` rA.
    //
    // rs: Sub. substitutes a pattern with the replacement. `pat` `src` `repl` rs. `repl` can
    //     be a block.
    // rt/rS: Sub with flags. `pat` `src` `repl` `flags` rS.
    //
    // Note: for flags, please refer to `godoc regexp/syntax`// >>>>
    addOp(&Op{"rm",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            pat, str := to_s(a), to_s(b)
            matched, err := regexp.MatchString("^" + pat + "$", str)
            if err != nil {
                panic(err.Error())
            }
            return wraps(to_i(matched))
        }},
    }})// >>>>
    addOp(&Op{"rf",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            pat, str := to_s(a), to_s(b)
            return wraps(to_i(to_bool(regexp.MustCompile(pat).FindString(str))))
        }},
    }})// >>>>
    addOp(&Op{"rr",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            pat, str := to_s(a), to_s(b)
            return wraps(regexp.MustCompile(pat).FindString(str))
        }},
    }})// >>>>
    addOp(&Op{"rl",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            pat, str := to_s(a), to_s(b)
            _pat := regexp.MustCompile(pat)
            indexes := _pat.FindStringIndex(str)
            return wraps(wrapa(_pat.FindString(str), indexes[0], indexes[1]))
        }},
    }})// >>>>
    addOp(&Op{"ra",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            pat, str := to_s(a), to_s(b)
            _pat := regexp.MustCompile(pat)
            matches := _pat.FindAllString(str, -1)
            if matches == nil {
                return wraps(wrapa())
            }
            res := wrapa()
            for _, m := range matches {
                res = append(res, m)
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"re",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            pat, str := to_s(a), to_s(b)
            _pat := regexp.MustCompile(pat)
            matches := _pat.FindAllString(str, -1)
            indexes := _pat.FindAllStringIndex(str, -1)
            if matches == nil {
                return wraps(wrapa())
            }
            res := wrapa()
            for i := 0; i < len(matches); i ++ {
                inds := indexes[i]
                res = append(res, wrapa(matches[i], inds[0], inds[1]))
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"ru",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            pat, str, flags := to_s(a), to_s(b), to_s(c)
            matched := regexp.MustCompile(fmt.Sprintf("(?%s)%s", flags, pat)).FindString(str)
            return wraps(to_i(to_bool(matched)))
        }},
    }})// >>>>
    addOp(&Op{"rs",// <<<<
    []TypedFunc{
        {"ssa", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c := x.Get3()
            pat, str := to_s(a), to_s(b)
            _pat := regexp.MustCompile(pat)
            if is_b(c) {
                return wraps(_pat.ReplaceAllStringFunc(str, func(s string) string {
                    env.Push(s)
                    to_b(c).Run(env)
                    return to_s(env.Pop())
                }))
            }
            return wraps(_pat.ReplaceAllString(str, to_s(c)))
        }},
    }})// >>>>
    addOp(&Op{"rt",// <<<<
    []TypedFunc{
        {"ssas", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b, c, d := x.Get4()
            pat, str, flags := to_s(a), to_s(b), to_s(d)
            _pat := regexp.MustCompile(fmt.Sprintf("(?%s)%s", flags, pat))
            if is_b(c) {
                return wraps(_pat.ReplaceAllStringFunc(str, func(s string) string {
                    env.Push(s)
                    to_b(c).Run(env)
                    return to_s(env.Pop())
                }))
            }
            return wraps(_pat.ReplaceAllString(str, to_s(c)))
        }},
    }})// >>>>
    // >>>>
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
            return nil
        }},
    }})// >>>>
    // The `Xfer` module// <<<<
    // Brief:// <<<<
    // xp: HTTP POST
    // xg: HTTP GET
    // xj: JSON encode
    // xk/xJ: JSON decode
    // xx: XML encode
    // xy/xX: XML decode// >>>>
    addOp(&Op{"xg",// <<<<
    []TypedFunc{
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            url := to_s(x.Get1())
            if ! strings.Contains(url, "://") {
                url = "http://" + url
            }
            resp, err := http.Get(url)
            if err != nil {
                panic(err.Error())
            }
            data, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                panic(err.Error())
            }
            return wraps(string(data))
        }},
    }})// >>>>
    addOp(&Op{"xp",// <<<<
    []TypedFunc{
        {"ss", 0x10,
        func(env * Environ, x * Stack) *Stack {
            a, b := x.Get2()
            url, body := to_s(a), to_s(b)
            if ! strings.Contains(url, "://") {
                url = "http://" + url
            }
            resp, err := http.Post(url, "", strings.NewReader(body))
            if err != nil {
                panic(err.Error())
            }
            data, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                panic(err.Error())
            }
            return wraps(string(data))
        }},
    }})// >>>>
    addOp(&Op{"xj",// <<<<
    []TypedFunc{
        {"v", 0x10,
        func(env * Environ, x * Stack) *Stack {
            data, err := json.Marshal(x.Get1())
            if err != nil {
                panic(err.Error())
            }
            return wraps(string(data))
        }},
    }})// >>>>
    addOp(&Op{"xk",// <<<<
    []TypedFunc{
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            var res interface{}
            err := json.Unmarshal([]byte(to_s(x.Get1())), &res)
            if err != nil {
                panic(err.Error())
            }
            return wraps(res)
        }},
    }})// >>>>
    addOp(&Op{"xx",// <<<<
    []TypedFunc{
        {"v", 0x10,
        func(env * Environ, x * Stack) *Stack {
            data, err := xml.Marshal(x.Get1())
            if err != nil {
                panic(err.Error())
            }
            return wraps(string(data))
        }},
    }})// >>>>
    addOp(&Op{"xy",// <<<<
    []TypedFunc{
        {"s", 0x10,
        func(env * Environ, x * Stack) *Stack {
            var res interface{}
            err := xml.Unmarshal([]byte(to_s(x.Get1())), &res)
            if err != nil {
                panic(err.Error())
            }
            return wraps(res)
        }},
    }})// >>>> // >>>>
    addOp(&Op{"y",// <<<<
    []TypedFunc{
        // Recurse function (No signature because it's not fixed)
        {"", 0x10,
        func(env * Environ, x * Stack) *Stack {
            defer func() {
                // Handle the index err paniced by the `Get` and `Pop` functions
                if err := recover(); err != nil {
                    if strings.Contains(to_s(err), "bounds") {
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
    addOp(&Op{"xt",// <<<<
    []TypedFunc{
        {"a", 0x10,
        func(env * Environ, x * Stack) *Stack {
            fmt.Printf("\x1b[33mDEBUG\x1b[0m type: %s\n", fulltype(x.Get1()))
            return nil
        }},
    }})// >>>>
// >>>>
}
