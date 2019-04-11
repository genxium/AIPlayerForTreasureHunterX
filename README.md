If you have trouble downloading the dependencies, e.g. `golang.org/x/sys` shown by `go get -u`, please try 

```
user@~> mkdir -p $GOPATH/src/golang.org/x/
user@~> cd $GOPATH/src/golang.org/x/
user@~> git clone https://github.com/golang/sys.git
```

, then specify local dependency of `golang.org/x/sys` in `go.mod` and rebuild the project.

Keep using `go-isatty v0.0.3` in `go.mod` if you don't want any trouble circumventing the GFW.
