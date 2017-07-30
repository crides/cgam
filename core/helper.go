package cgam

import (
    "fmt"
    "math"
    "os/exec"
)

// Helper functions
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
    t := s[x:]
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
        size := int(math.Min(float64(as), float64(bs)))
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

func flatten(a []interface{}) []interface{} {
    res := wrapa()
    for _, o := range a {
        if ! is_l(o) {
            res = append(res, o)
        } else {
            res = append(res, flatten(to_l(o))...)
        }
    }
    return res
}

func carte_product(al, bl []interface{}, flat bool) []interface{} {
    res := wrapa()
    if flat {
        for _, i := range al {
            for _, j := range bl {
                res = append(res, wrapa(i, j))
            }
        }
    } else {
        for _, i := range al {
            for _, j := range bl {
                res = append(res, append(to_l(i), j))
            }
        }
    }
    return res
}

func Range(n int) []interface{} {
    res := wrapa()
    for i := 0; i < n; i ++ {
        res = append(res, i)
    }
    return res
}

func round(x, unit float64) float64 {
    if x > 0 {
        return float64(int64(x / unit + 0.5)) * unit
    }
    return float64(int64(x / unit - 0.5)) * unit
}

