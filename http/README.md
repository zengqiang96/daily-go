`het/http`包作为Go标准库的一部分，提供了HTTP客户端和服务端的实现。使用Go来构建web服务不再需要其他的第三方包

### 简单的web server

```go
import "net/http"

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	http.ListenAndServe(":3000", mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Gophers!"))
}
```

`mux := http.NewServeMux()`定义了一个多路复用器(multiplexer)，用来路由进来的请求到相应的handler

### 使用Server启动web server

除了上面的那种方法启动web server，还可以使用Server来启动

```go
// from https://golang.org/src/net/http/server.go?s=77156:81268#L2480
// A Server defines parameters for running an HTTP server.

// The zero value for Server is a valid configuration.
type Server struct {
	Addr    string  // TCP address to listen on, ":http" if empty
	Handler Handler // handler to invoke, http.DefaultServeMux if nil
	TLSConfig *tls.Config
	ReadTimeout time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout time.Duration
	IdleTimeout time.Duration
	MaxHeaderBytes int
	TLSNextProto map[string]func(*Server, *tls.Conn, Handler)
	ConnState func(net.Conn, ConnState)
	ErrorLog *log.Logger
	BaseContext func(net.Listener) context.Context
	ConnContext func(ctx context.Context, c net.Conn) context.Context
	disableKeepAlives int32     // accessed atomically.
	inShutdown        int32     // accessed atomically (non-zero means we're in Shutdown)
	nextProtoOnce     sync.Once // guards setupHTTP2_* init
	nextProtoErr      error     // result of http2.ConfigureServer if used
	mu         sync.Mutex
	listeners  map[*net.Listener]struct{}
	activeConn map[*conn]struct{}
	doneChan   chan struct{}
	onShutdown []func()
}
```

使用server重新定义一个web server

```go
import "net/http"

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	httpServer := http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	httpServer.ListenAndServe()
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Gopher!"))
}
```

### multiplexer

multiplexer用来定义web应用的路由


main.go
```go
import "net/http"

func main()  {
	registerRoutes()
	httpServer := http.Server{
		Addr: ":3000",
		Handler: mux,
	}

	httpServer.ListenAndServe()
}
```

routes.go

```go
import "net/http"

var mux = http.NewServeMux()

func registerRoutes()  {
	mux.HandleFunc("/home", home)
	mux.HandleFunc("/about", about)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
}
```

handler.go

```go
import "net/http"

func logout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("logout route"))
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("login route"))
}

func about(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("about route"))

}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("home route"))
}
```

### handler

handler负责处理进来的http请求，在multiplexer中可以定义很多不同的路由

有2中方法来定义handler

使用struct，为了让struct是有效的handler，必须满足handler接口的定义

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

首先定义一个http handler的struct

```go
type CustomHandler struct {}
```
然后实现Handler接口
```go
func (c CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("custom handler!"))
}
```
使用`CustomHandler`

```go
type CustomHandler struct {}

func (c *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("custom handler!"))
}

func main()  {
	handler := CustomHandler{}
	mux := http.NewServeMux()
	mux.Handle("/", &handler)
	http.ListenAndServe(":3000", mux)
}
```

也可以使用function作为handler，这多亏了`HandlerFunc`

```go
// source : https://golang.org/src/net/http/server.go?s=61509:61556#L1993
// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

* `HandlerFunc`相当是适配器，它接受一个函数，返回一个实现了`ServeHTTP`方法的`Handler`


```go
import "net/http"

func functionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("function as http handler"))
}

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", functionHandler)
	http.ListenAndServe(":3000", mux)
}
```

### http request

http请求在Go中使用Request结构体来表示

```go
// source : https://golang.org/src/net/http/request.go?s=3252:11812#L97
type Request struct {
	Method string
	URL *url.URL
	Proto      string // "HTTP/1.0"
	ProtoMajor int    // 1
	ProtoMinor int    // 0
	Header Header
	Body io.ReadCloser
	GetBody func() (io.ReadCloser, error)
	ContentLength int64
	TransferEncoding []string
	Close bool
	Host string
	Form url.Values
	PostForm url.Values
	MultipartForm *multipart.Form
	Trailer Header
	RemoteAddr string
	RequestURI string
	TLS *tls.ConnectionState
	Cancel <-chan struct{}
	Response *Response
	ctx context.Context
}
```

我们从Request中获取很多信息

* Request的Body
* Query参数
* http Headers
* Post表单

```go
import (
	"net/http"
	"strconv"
)

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", requestInspection)
	http.ListenAndServe(":3000", mux)
}

func requestInspection(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Method: " + r.Method + "\n"))
	w.Write([]byte("Protocol Version: " + r.Proto + "\n"))
	w.Write([]byte("Host: " + r.Host + "\n"))
	w.Write([]byte("Referer: " + r.Referer() + "\n"))
	w.Write([]byte("User Agent: " + r.UserAgent() + "\n"))
	w.Write([]byte("Remote Addr: " + r.RemoteAddr + "\n"))
	w.Write([]byte("Requested URL: " + r.RequestURI + "\n"))
	w.Write([]byte("Content Length: " + strconv.FormatInt(r.ContentLength, 10) + "\n"))

	for k, v := range r.URL.Query(){
		w.Write([]byte("Query string: key=" + k + " value=" + v[0] + "\n"))
	}
}
```

### http response

http response和http request类似，它代表http request的响应。由`Response`结构体定义

```go
// source : https://golang.org/src/net/http/response.go?s=731:4298#L25
type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0
	Header Header
	Body io.ReadCloser
	ContentLength int64
	TransferEncoding []string
	Close bool
	Uncompressed bool
	Trailer Header
	Request *Request
	TLS *tls.ConnectionState
}
```

我们并不是直接使用`Response`结构体。而是使用`RequestWriter`接口来构建http response。`ResponseWriter`定义如下

```go
// source : https://golang.org/src/net/http/server.go?s=2985:5848#L84
type ResponseWriter interface {
	Header() Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}
```

> 注意：如果返回的status code不是200，在调用w.Writer()之前必须调用w.WriteHeader()

```go
import "net/http"

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", unauthorized)

	http.ListenAndServe(":3000", mux)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("you do not have permission to access this resource.\n"))
}
```

访问结果

```text
$curl http://localhost:3000/home\?name\=sumit
you do not have permission to access this resource.

$curl -I http://localhost:3000/home
HTTP/1.1 401 Unauthorized
Date: Thu, 28 May 2020 12:53:10 GMT
Content-Length: 52
Content-Type: text/plain; charset=utf-8
```

### http headers

Go定义了`Header`来表示http headers

```go
type Header map[string][]string
```

前面我们可以看到在`Request`和`Response`结构体中都有header部分

我们来看一下`Request`中的header

```text
curl http://localhost:3000
map[Accept:[*/*] User-Agent:[curl/7.64.1]]
```

浏览器输出

```text
map[Accept:[text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9] Accept-Encoding:[gzip, deflate, br] Accept-Language:[zh-CN,zh;q=0.9,en;q=0.8] Connection:[keep-alive] Cookie:[Webstorm-bc417914=9eec914e-fe08-47c9-b0ac-9475c286e9bd; experimentation_subject_id=IjgwMDNjNDVhLTMzZmItNDUyYy1iN2EzLWM0NWJmNmZkYjU2MiI%3D--00c061ec68d8bc350bcc17db6667cce85b374228; sidebar_collapsed=false] Sec-Fetch-Dest:[document] Sec-Fetch-Mode:[navigate] Sec-Fetch-Site:[none] Sec-Fetch-User:[?1] Upgrade-Insecure-Requests:[1] User-Agent:[Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36]]
```

### 获取header内容

有2种方法从header种获取内容

1.直接使用Header的Get方法，回获取string的结果

```go
header := r.Header
accept1 := header.Get("Accept")
```

2.直接使用map的访问方法获取

```go
header := r.Header
accept2 := header["Accept"]
```

根据不同的需求，我们可以选择不同的方法来处理

### 设置header内容

在构建Response时也可以设置header，我们可以使用`Set`方法

让我们来设置一下Header `ALLOWED`

```go
import "net/http"

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", setHeader)
	http.ListenAndServe(":3000", mux)
}

func setHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ALLOWED",  "GET,POST")
	w.Write([]byte("set allowed headers\n"))
}
```

输出结果

```text
$curl -i http://localhost:3000
HTTP/1.1 200 OK
Allowed: GET,POST
Date: Thu, 28 May 2020 15:10:56 GMT
Content-Length: 20
Content-Type: text/plain; charset=utf-8

set allowed headers
```

> 注意：不像`ResponseWriter`的`WriteHeader()`方法并不是用来设置`Response`的header，net/http包中对这部分有说明

```go
// WriteHeader sends an HTTP response header with the provided
// status code.
WriteHeader(statusCode int)
```

### WriteHeader()的使用

`WriteHeader`用来设置http的status code，`Write`方法会在写数据之前调用`WriteHeader(StatusOK)`，如果返回的http status不是200，那么在调用`Write`之前调用`WriteHeader`就很重要了

```go
if !w.wroteHeader {
  w.WriteHeader(StatusOK)
}
```

正确的用法

```go
import "net/http"

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", setHeader)
	http.ListenAndServe(":3000", mux)
}

func setHeader(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad request!\n"))
}
```

错误的用户

```go
import "net/http"

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", setHeader)
	http.ListenAndServe(":3000", mux)
}

func setHeader(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Bad request!\n")) // 这将会设置status为 200 ok
	w.WriteHeader(http.StatusBadRequest) // 这里的WriteHeader设置将无效
}
```

来看看这样的输出结果，返回的http status仍然是 200 OK

```go
$curl -i http://localhost:3000
HTTP/1.1 200 OK
Date: Thu, 28 May 2020 15:42:01 GMT
Content-Length: 13
Content-Type: text/plain; charset=utf-8

Bad request!
```

### 查询字符串

从request中获取查询字符串是最常见的一种场景

在`Request`中有`URL`字段

```go
type Request struct {
  // other field omitted
  URL *url.URL
}
```

`URL`有自己的方法获取查询字符串

```go
type URL struct {
  // fields omitted
  RawQuery   string    // encoded query values, without '?'
}

// Query parses RawQuery and returns the corresponding values.
// It silently discards malformed value pairs.
// To check errors use ParseQuery.
func (u *URL) Query() Values {
	v, _ := ParseQuery(u.RawQuery)
	return v
}
```

可以调用这个`Query()`方法获取查询字符串

```go
import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", showQuery)
	http.ListenAndServe(":3000", mux)
}

func showQuery(w http.ResponseWriter, r *http.Request) {
	queryString := r.URL.Query()
	w.Write([]byte("query strings are\n"))
	w.Write([]byte("Name:" + queryString.Get("name") + "\n"))
	w.Write([]byte("Email:" + queryString.Get("email") + "\n"))
}
```

查询结果

```text
curl -G -d 'name=test' -d 'email=test@gmail.com' localhost:3000
query strings are
Name:test
Email:test@gmail.com
```

### 使用cookies

因为http是无状态的，所以有多种方案来维持请求之间的数据，其中一种就是cookie。

> HTTP Cookie（也叫Web Cookie或浏览器Cookie）是服务器发送到用户浏览器并保存在本地的一小块数据，它会在浏览器下次向同一服务器再发起请求时被携带并发送到服务器上。通常，它用于告知服务端两个请求是否来自同一浏览器，如保持用户的登录状态。Cookie使基于无状态的HTTP协议记录稳定的状态信息成为了可能。 --- MDN web docs

`net/http`包中定义的cookie结构如下

```go
// source : https://github.com/golang/go/blob/go1.13.7/src/net/http/cookie.go#L19
type Cookie struct {
	Name  string
	Value string

	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}
```

### 设置cookie

为了设置cookie，服务器需要发送`Set-Cookie`header。浏览器会根据header来创建cookie。Go提供了`SetCookie`函数来设置cookie，这个函数需要`ResponseWriter`和`Cookie`参数

```go
//source : https://golang.org/src/net/http/cookie.go?s=3954:4002#L150
// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func SetCookie(w ResponseWriter, cookie *Cookie) {
	if v := cookie.String(); v != "" {
		w.Header().Add("Set-Cookie", v)
	}
}
```

尝试设置一下

```go
import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", setCookies)
	http.ListenAndServe(":3000", mux)
}

func setCookies(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name: "cookie-1",
		Value: "hello world",
	}
	http.SetCookie(w, &cookie)
}
```

请求一下

```text
$curl -i http://localhost:3000
HTTP/1.1 200 OK
Set-Cookie: cookie-1="hello world"
Date: Sun, 31 May 2020 12:43:17 GMT
Content-Length: 0
```

### 获取Cookie

如果之前的请求设置了cookie，浏览器会将cookie再发送回server。`Request`有`Cookies()`方法来获取请求中的cookies。

```go
// source : https://github.com/golang/go/blob/go1.13.7/src/net/http/request.go#L408
// Cookies parses and returns the HTTP cookies sent with the request.
func (r *Request) Cookies() []*Cookie {
	return readCookies(r.Header, "")
}
```

除此之外还可以使用`Cookie`方法通过name来获取单个cookie。

```go
//source : https://github.com/golang/go/blob/go1.13.7/src/net/http/request.go#L419
// Cookie returns the named cookie provided in the request or
// ErrNoCookie if not found.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (r *Request) Cookie(name string) (*Cookie, error) {
	for _, c := range readCookies(r.Header, name) {
		return c, nil
	}
	return nil, ErrNoCookie
}
```

看看获取cookie的方式

```go
import (
	"fmt"
	"net/http"
)

func main()  {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getCookies)
	http.ListenAndServe(":3000", mux)
}

func getCookies(w http.ResponseWriter, r *http.Request) {
	cookies := r.Cookies()
	for _, cookie := range cookies{
		fmt.Fprintln(w, cookie)
	}

	cookie, err := r.Cookie("cookie-1")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
	fmt.Fprintln(w, cookie)
}
```

执行一下

```text
$curl -b 'cookie-1=hello' http://localhost:3000
cookie-1=hello
cookie-1=hello
```

### 使用Sessions

Session是另外一种持久化数据的技术。Sessions常常是将状态存储在服务端，和cookie将状态存储在浏览器相反。因为cookie存储在用户的浏览器中，它们通常用于增强用户体验。

另一方面，session在服务端进行管理，所有我们有更多的控制权和安全性。Session广泛用于敏感信息，例如用户是否已登录系统。

Session通常自己实现很困难，我们经常使用一些第三方的session管理包

### 获取和设置session数据

使用`github.com/gorilla/sessions`来管理session

这里使用CookieStore来存储session数据，也可以使用文件系统或数据库等

```go
import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
)

// for prod use secure key, not hard-coded
var store = sessions.NewCookieStore([]byte("1234"))

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", sessionHandler)
	http.ListenAndServe(":3000", mux)
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "custom-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	val := session.Values["hello"]
	if val != nil {
		fmt.Fprintln(w, "retrieving existing session: ")
		fmt.Fprintln(w, val)
		return
	}
	session.Values["hello"] = "world"
	session.Save(r, w)
	fmt.Fprintln(w, "no existing session found, set value for session")
}
```


