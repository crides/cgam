Cgam
----

# Introduction

Cgam is a stack-based programming language implemented in Go that can be used for code golfing. It's derived from [Cjam](https://sourceforge.net/projects/cjam/), but has some major differences.

## Stack-based languages

Stack-based languages works on a stack that grows and shrinks. The operations available for these languages are commonly pushing, poping or operating on one or more values. But languages like cjam (and of course, cgam) provide more operations for operating on the stack and also on values. Stack-based languages usually uses the [Reverse-Polish notation](https://en.wikipedia.org/wiki/Reverse_Polish_notation) syntax, which eliminates the need for braces and making the syntax much more simpler.

In Cjam or Cgam, the common operations would be pushing a value, rotating the stack, calling a function (or procedure, in a more exact way), popping a value, etc.

## A brief intro for Cjam

In Cjam, there are 26 variables for each upper case letters other than the stack. It uses predefined operators to operate on the stack and the values. It has 5 basic types: integers, doubles, characters, lists (and strings, which are lists of characters), and blocks. Blocks are actually procedures, which consists of many operators. Value literals and variables are pushed onto the stack. If a variable is a block, it is executed when pushed. Other than the Reverse-Polish Notation, Cjam also has infix operators, which has help from the parser to help them complete more tasks.

## Difference from Cjam
1. Cjam uses `;` to pop a value from the stack. Cgam Use `;` for line comment
2. Cjam uses `o` to output a string representation of a value (`str()` in Python, not `repr()`) and no newline, and `p` to print the value representation (`repr()` this time) and a newline. Cgam uses `o` for the `os` module, and `p` to function the same as the Cjam `o` operator.
3. Cjam uses <code>&#96;</code> for converting a value to its value representation string (`repr()`), but in Cgam, it is used for popping a value from the stack, and `v` is used for a value representation.
4. `.` in Cjam is used on two lists like:
```python
    [a[i:i+1] *op* b[i:i+1] for i in range(max(len(a), len(b)))]
```
and `:` is used to assign a value to a variable, or do `map` or `fold` operations on a list.
In Cgam, this 2 are switched, because one-list operations should be used on one dot, and two-list operations should be used on two dots (colon). And like `.`, `:` can also be followed by a block literal and used for folding (not mapping, because it can be achieved by `%`, and the lengths are the same).

5. In Cgam, `e[` and `e]` are switched.
6. In Cgam, `e<`, `e>` and `m<`, `m>` are switched. Min and max should be in `math` module, and bit shifting and list rotating should be in `extended` module (by the way, the `extended` module also contains many list operations).
7. In Cjam, `j` (recurse functions) can have a pre-initialized cache. In Cgam, `j` moved to `y`, and the init cache was removed, but cache is still used in the evaluation.
8. In Cgam, variables can have names longer than one character.
9. In Cgam, more modules (or extended operators) are added (`os`, `regex`, `xfer`).
10. In Cgam, user defined modules can be added.
11. In Cgam, name spaces are added to simplify invoking the extended operators.

**NOTE:** Some features above are still not available at the moment, as the project is still IN DEVELOPMENT!

## Some rationales for making the changes
1. Why long variable names?

Well, this may be taken from [Paradoc](https://github.com/betaveros/paradoc/), another stack-based language derived from Cjam, but long-named variables make the program much easier to understand, just like when you compare a compressed js file and an uncompressed one. Among other things, Cgam can be a easy way to jot down the code, but also keeping the original variable names.

2. Why more modules?

I don't even know exactly why myself ... But somehow I want to make it more like a general programming language but not just for some simple functions. Anyway, those modules shouldn't bother your code golfing!

3. Why use ... Go?

Sorry, I just want to practise myself in using this language. Go is a faster language compare to Java (the startup for Cjam is slow), but it also has wrapped types and operations compare to C.

4. Also, maybe reverse to what Paradoc want to do, I don't want to include any characters outside of ASCII. Other than making the program shorter by one or two bytes isn't that important, but the program should also be easily typed on a normal 104/105 key keyboard.

5. But similar to Paradoc, I do want to make cgam easier to write, that the operators and their meanings *must* be related. So I changed some operators' names (especially the alphabet operators), to reduce weird operators (like `j`, `h`, and `g`, so I replaced `j` with `y`, because the `Y` combinator is used for recursion in lambda calculus).

# Building and usage

Nothing special, just clone the code and `go build cgam.go`! (Sorry, I haven't familiar myself with `go get` and file splitting in Go ...) Running `./cgam` will then take you to the REPL. (Running file and CLI arguments will soon be supported.)

# Contributing

I am open to pull requests, and you can help me test and open issues! And I have some concern on which features should I choose to use in Cgam too. You can check out the [discussion](#discussion) section below.

## Discussion

**NOTE:** This sections includes some thoughts I have about which feature to use or not.

1. **name spaces**

To add more functionalities to Cgam, I used many two-character operators, but that immediately increases the program length. To make it more "code-golfable", I want to add name spaces so that through `u?` (`?` is a char), later in the source code we don't have to write two characters. For example, if somewhere we used `uo`, which means switching to the `os` name space (or module), then before the next `u?`, we can use `x` to directly execute a command instead of `ox`.

But that leads to another problem. This feature is definitely in the parser. Using the example from above, when I use `uo` and then `x`, how can the parser know if the `x` is the `ox` operator, or is it the first character in something like `xg`? Of course I can look forward a char to see if the two character can form a new operator or not, but that definitely decreases readability.

2. **Explicit variable pushing**

Another concern I have is whether to add a `.X` operator to explicitly push the variable onto the stack. The reason is, sometimes a two-char operator's second char is upper case, and if I use name spaces to abbreviate it, then it may become a upper case operator, which can be confusing to the parser and programmer. So maybe a explicit variable pushing operator should be added, and then we can also use more operators (we can use uppercase letters in single-char operators).

3. **User-defined modules**

This is a little bit far-streching to a language like Cgam, but it is interesting to add such a feature, though it may be difficult. The problems are:

- How to recognize a module file? Maybe add a line like this: (somehow like shebang ...)?
```lisp
        ; Package x
```
- How to import a module? Still use `ux`?
