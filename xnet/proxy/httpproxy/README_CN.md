# httpproxy
#### ([EN](./README.md)/CN)
## 说明
- 这是http的简单代理实现，内部使用了net/http包的实现，所以只支持http1类型的协议，http2并不支持（http2更复杂，个人没有足够的精力去处理）。
- 至于为什么会有这玩意？为什么不用net/http的proxy？你会这么想是正常的，对于我来说，http.server存在一些臃肿，对于我但场景是处理流连接但场景，使用这个更轻量。
- 这个实现经过了实际在多个浏览器的测试，并没有产生影响使用的错误。
## 原理
- 解析http报文，得到目标的地址，连接然后建立io.copy（无论是长连接还是短连接，建立的流连接是不会改变的，这里采用的io.copy也是因为不关心后续报文的内容，解析内容只是为了获取目标地址，客户端是否复用连接是客户端该考虑的事情）。
- 注意，因为读取了第一个报文内容，对于报文的方法的意义是不一样的，以下会分开说明：
  - other：因为已经从连接读取了内容，这是实实在在的有效载荷，需要做的就是将报文存储起来，在建立连接后，将该报文发送出去。
  - CONNECT：相比http，https有一个握手的过程，这个过程代理方不应该参与（为了通信加密安全），客户端会先发送一个CONNECT的方法，用来确认代理与服务端是否建立了连接，当建立连接后，需要返回“200 Connection Established”的内容，后续就是客户端与服务端的通信内容（例如https的握手）。
  - 总结一下就是，当客户端发送了CONNECT，意味着客户端是将第一个报文作为握手信息，是不具有有效载荷（只是获取到目标地址），这类场景大多是在https的情况。
- 读取到的地址，当不携带端口信息时，默认时80端口。
## 使用
- ```go
  package main
  
  import "github.com/peakedshout/go-pandorasbox/xnet/proxy/httpproxy"
  
  func main() {
    server, _ := httpproxy.NewServer(&httpproxy.ServerConfig{
      ReqAuthCb:   nil, // proxy auth callback (e.g. username and password)
      Forward:     nil, // unique connections created (e.g. relay to proxy)
      DialTimeout: 0,
    })
    defer server.Close()
    _ = server.ListenAndServe("tcp", "0.0.0.0:80")
  }
  ```
- 更多请看test文件。