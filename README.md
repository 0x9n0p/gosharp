# The GoSharp Programming Language

GoSharp is an open-source programming language forked from [Golang](https://github.com/golang/go).

## Features

### Immediate Return https://github.com/0x9n0p/gosharp/pull/1
Imagine we have a function named 'Callee' that returns an error. If we want to handle the error, we need to use an if statement to check the error and return it if that isn't nil. 
It's simple and is one of the reasons that we love Golang! But, at least 3 lines of code will be added to our caller function's body. <br>
With this feature, instead of using if statements, we can use a question mark after the right parenthesis!
```go
func caller() error {
  callee()?
  callee2()?
  return nil
}
```


## Install From Source

Make sure you have Golang installed.

```
$ go version
go version go1.21.6 linux/amd64
```

Clone source code and build the entire project

```bash
git clone https://github.com/0x9n0p/gosharp.git && cd gosharp/src && ./make.bash
```

Export bin directory to find the compiled tools

```bash
export PATH="$GOPATH/src/gosharp/bin:$PATH"
```

And now, if you run the version subcommand, you must see the gosharp version too.

```
$ go version
gosharp version 1.0-nightly
go version devel go1.23-0ba426291c Thu Jun 6 14:02:00 2024 +0330 linux/amd64
```

## Contributing

Feel free to send Pull Requests, fix typos, grammatical mistakes. We appreciate your help!
