If you have trouble downloading the dependencies, e.g. `golang.org/x/sys` shown by `go get -u`, please try 

```
user@~> mkdir -p $GOPATH/src/golang.org/x/
user@~> cd $GOPATH/src/golang.org/x/
user@~> git clone https://github.com/golang/sys.git
```

and rebuild the project.

