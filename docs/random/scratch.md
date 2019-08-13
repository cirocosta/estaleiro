# from scratch

## context

As the best way of not being exposed to CVEs is not having a single thing
installed, `scratch` allows one to have a totally blank layer that you can
build upon.

That's very useful for those who just build a binary and no extra dependencies
are needed.

```

  (scratch)
     .---------.                        .--------.
     |         |			|        |
     | nothing |  + your binary  ==>    | binary |
     |         |                        |        |
     *---------*                        *--------*


```

While it's possible to have fully static binaries that are self contained,
that's not always the case (e.g., binaries that rely on dynamic linking
`glibc`).

For those cases, it's useful to have the contents of certain packages (like,
`libc` and `ca-certificates`).


## the proposal

We could learn from what the [distroless][distroless] folks have been doing and
incorporate that into the way that `scratch`-based images could be built.

As such, that'd mean:


1. allowing `base_image { name = "scratch" }` being specified
  
under the hood, it'd start from an `llb.Scratch()` state rather than an
`llb.Image()`


2. allowing the ability to pick up files from packages while keeping the
trace of where those came from

would allow one to manually pick libc/ca-certificates stuff


## implementation

initially, leveraging `scratch` would mean not being able to leverage `apt`,
thus, we should perform that check during the validation phase.

> should we just tie a `Validate()` method to `config`? :thinking:


[distroless]: https://github.com/GoogleContainerTools/distroless

