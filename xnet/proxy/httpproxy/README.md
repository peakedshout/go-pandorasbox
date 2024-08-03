# httpproxy
#### (EN/[CN](./README_CN.md))
## Explanation
- This is a simple proxy implementation of http. It uses the implementation of the net/http package internally, so it only supports http1 type protocols, and http2 does not support it (http2 is more complicated, and I personally don’t have enough energy to deal with it).
- As for why this thing exists? Why not use net/http proxy? It's normal for you to think so. For me, http.server is a bit bloated. For me, the scenario is to handle streaming connections, so using this is more lightweight.
- This implementation has been tested in multiple browsers and has not produced errors that affect its use.
## Theory
- Parse the http message, get the target address, connect and then establish io.copy (whether it is a long connection or a short connection, the established stream connection will not change. The io.copy used here is also because it does not care about the content of subsequent messages. , parsing the content is only to obtain the target address. Whether the client can reconnect is something the client should consider).
- Note that because the content of the first message is read, the meaning of the message method is different, which will be explained separately below:
  - other: Because the content has been read from the connection, this is a real payload. All that needs to be done is to store the message and send the message after the connection is established.
  - CONNECT: Compared with http, https has a handshake process. The agent should not participate in this process (for the sake of communication encryption security). The client will first send a CONNECT method to confirm whether the agent and the server have established a connection. When the connection is established, the content of "200 Connection Established" needs to be returned, followed by the communication content between the client and the server (such as the https handshake).
  - To summarize, when the client sends CONNECT, it means that the client uses the first message as handshake information and does not have a payload (just obtains the target address). Most of these scenarios are in https.
- The read address, when it does not carry port information, defaults to port 80.
## Use
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
- Please see the test file for more information.

