package cgam

import (
    "fmt"
    "strconv"
)

// Type predicates//
func typeof(a interface{}) string {
    switch a.(type) {
    case string:
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
    return typeof(a) == "double" || typeof(b) == "double"
}
//
// Type conversions
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
    case string:
        as := make([]interface{}, 0)
        for _, i := range b {
            as = append(as, wrap(i))
        }
        return as
    default:
        // When encountering single values, it just wraps it up; for integer ranges, `Range()`
        // is used
        return wrapa(b)
    }
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
        return fmt.Sprint(b)    // To simplify string conversion
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
    }
    panic(fmt.Sprintf("Invalid type %s!", fulltype(a)))
}

func to_bool(a interface{}) bool {
    switch b := a.(type) {
    case int, rune:
        return b != 0
    case float64:
        return b != 0
    case []interface{}:
        return len(b) != 0
    case string:
        return len(b) != 0
    default:
        not_convertible(a, "bool")
    }
    return false
}
//
