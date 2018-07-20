Cgam
----

# Introduction

Cgam is a stack-based programming language implemented in Go that can be used for code golfing. It's derived from [CJam](https://sourceforge.net/projects/cjam/), but has some major differences.

## Stack-based languages

Stack-based languages work on a stack that can grow and shrink. Operations available for these languages commonly include pushing, popping and operating on one or more values, but languages like CJam (and of course, Cgam) provide more operations for operating on the stack and its values. Stack-based languages usually use [reverse Polish notation](https://en.wikipedia.org/wiki/Reverse_Polish_notation) which eliminates the need for braces and making the syntax much simpler.

In CJam and Cgam, the common operations would be pushing a value, rotating the stack, calling a function (or, more accurately, a procedure), popping a value, etc.

## A brief intro to CJam

In CJam, there are 26 variables, one for each upper case letter, which can be used for storage outside of the stack. It uses predefined operators to operate on the stack and the values. It has 5 basic types: integers, doubles, characters, lists (and strings, which are lists of characters), and blocks. Blocks are actually procedures, which consists of many operators. Value literals and variables are pushed onto the stack. If a variable is a block, it is executed when pushed. Some operators are also syntactically overloaded as infix as well.

## Differences from CJam

1. Use `` ` `` instead of `;`, and `v` instead of `` ` ``. `;` denotes a line comment.
2. Use `p` instead of `o`. `o` is used for the `os` module. There is no equivalent of CJam's `p` in Cgam; you can use `voNo` instead.
3. CJam's `.` is used to zip two lists with an operator, and `:` is used to map or fold over a list (or assign to a variable). In Cgam, these two are switched, because a single dot `.` should denote an operation on a single list, and a double dot `:` should denote an operation on two lists. Cgam's `.` can also be followed by a block literal and used for folding (not mapping, because it can be achieved by `%`, and the lengths are the same).
4. The meanings of `e[` and `e]` meanings are switched from what they are in CJam.
5. The meanings of `e<`, `e>` and `m<`, `m>` are switched from what they are in CJam. Min and max operators belong in the `math` module, and bitshifts and list rotation should be in `extended` module (which contains many other list operations).
6. `y` is used instead of `j`, and does not take an initial cache (though a cache is still used in evaluation). This was done because `j` is a strange name, and `y` is more likely to evoke recursion to those familiar with lambda calculus.
7. Variables can have names longer than one character.
8. Cgam has the additional namespaces `o` (for `os`), `r` (for `regex`), and `x` (for `xfer`).
9. Cgam supports user-defined modules.
10. Cgam supports additional namespaces are added to simplify invoking the extended operators.

**NOTE:** Some features above are still not available at the moment, as the project is still IN DEVELOPMENT!

## Rationale

1. Why long variable names?

Well, this may be taken from [Paradoc](https://github.com/betaveros/paradoc/), another stack-based language derived from CJam, but long-named variables make the program much easier to understand, just like when you compare a compressed js file and an uncompressed one. Among other things, Cgam can be a easy way to jot down the code, but also keeping the original variable names.

2. Why more modules?

I don't even know exactly why myself... but somehow I want to make it more like a general programming language but not just for some simple functions. Anyway, those modules shouldn't bother your code golfing!

3. Why use Go?

Sorry, I just want to practise myself in using this language. Go is a faster language compare to Java (the startup for Cjam is slow), but it also has wrapped types and operations compare to C.

## Design Goals

Unlike Paradoc, I don't want to include any characters outside of ASCII. Other than making the program shorter by one or two bytes, it isn't that important, but the program should also be easily typed on a normal 104/105 key keyboard. Like Paradoc, however, I do want to make Cgam easier to write than CJam. In particular, operators's names should correlate with their meaning, which CJam fails at in several places (such as `j`, `g`, and `h`).

# Building

Nothing special, just clone the code and `go build cgam.go`! (Sorry, I haven't familiar myself with `go get` and file splitting in Go ...) Running `./cgam` will then take you to the REPL. (Running file and CLI arguments will soon be supported.)

# Contributing

I am open to pull requests, and you can help me test and open issues! And I have some concern on which features should I choose to use in Cgam too. You can check out the [discussion](#discussion) section below.

## Discussion

This section includes some thoughts I have about which feature to use or not.

1. **Namespaces**

To add more functionality to Cgam, I used many two-character operators, but that immediately increases the program length. To make it more "code-golfable", I want to add name spaces so that through `u?` (`?` is a char), later in the source code we don't have to write two characters. For example, if somewhere we used `uo`, which means switching to the `os` name space (or module), then before the next `u?`, we can use `x` to directly execute a command instead of `ox`.

But that leads to another problem. This feature is definitely in the parser. Using the example from above, when I use `uo` and then `x`, how can the parser know if the `x` is the `ox` operator, or is it the first character in something like `xg`? Of course I can look forward a char to see if the two character can form a new operator or not, but that definitely decreases readability.

2. **Explicit variable pushing**

Another concern I have is whether to add a `.X` operator to explicitly push the variable onto the stack. The reason is, sometimes a two-char operator's second char is upper case, and if I use name spaces to abbreviate it, then it may become a upper case operator, which can be confusing to the parser and programmer. So maybe a explicit variable pushing operator should be added, and then we can also use more operators (we can use uppercase letters in single-char operators).

3. **User-defined modules**

This is a little bit farfetched to a language like Cgam, but it is interesting to add such a feature, though it may be difficult. The problems are:

- How to recognize a module file? Maybe add a line like this: (somehow like shebang ...)?

```lisp
        ; Package x
```
- How to import a module? Still use `ux`?
