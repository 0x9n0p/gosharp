# The GoSharp Programming Language

GoSharp is an open-source programming language forked from [Golang](https://github.com/golang/go). This language will
have features that make code hard
to read and may explode the production server.

### Install From Source

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

### Contributing

Feel free to send Pull Requests, fix typos, grammatical mistakes. We appreciate your help!
