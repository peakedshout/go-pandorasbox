# socks
#### ([EN](./README.md)/CN)
## 说明
- 这是一个纯golang实现的socks4/4a/5的实现，参考了规范文档，整合抽象接口并支持更详细的配置，请注意，该库可操作的空间十分巨大，如果不明白自己在做什么，那么最好是什么也不做，采用默认的方式。
- 至于为什么会有这玩意？这很简单的，因为目前golang实现的socks并不能完好的满足我，并且这socks并不是十分困难的事情。
- 这个实现经过了多个socks客户端和服务端的测试，并没有产生影响使用的错误。
- 注意，这个实现是根据简短的规范文档所实现，因为文档并没有仔细指导细节，所以存在争议是正常的。
## 原理
- 在阅读原理之前，我认为没有什么是比阅读规范文档更直接的，以下是路标：
  - socks4/4a： not found rfc (or you tell me)
  - socks5：rfc1929  rfc1928
- socks4 只有CONNECT和BIND方法，并且是tcp流，用户验证的方式是userid，这个userid理解成sessionid或许更直接和通用（？）
- socks4a 是对 socks4 进行的一个域名补丁，因为socks4在设计的时候，并没有考虑ipv6和域名的问题，而socks4a也只是解决域名问题，相比那时ipv6还没有纳入考虑范围（？）
- socks5 相比前两者，socks5具有UDPASSOCIATE方法，这是一个支持udp代理转发的方法，并且socks5支持ipv6和更好的支持域名。
- 注意，因为规范文档的说明十分简略，在socks5的UDPASSOCIATE存在歧义，下面进行进一步的说明：
  - 有关这个服务端创建的udp套接字是否参与直接的消息接收？（也就是自身是否是接收端和发送端？）
    - 关于这一点，在规范文档中说明的十分模糊，关于socks的安全性，接收和发送进行套接字隔离，这似乎更符合socks的代理理念（事实上确实有实现是这么做的）
    - 但接收与发送的隔离，会产生更多的维护开销，这点需要看个人的取舍，本实现采用的是隔离方案。
  - 有关客户端并没有发送客户端的预期地址？（即全零地址，无效地址，这里指的是全零ip并且全零端口）
    - 关于这一点，在规范文档中说明的十分模糊，开发者实现的方案无非只有两种：
    - 一是取发送端的ip作为客户端的地址进行服务，这显然是最最简单的，但会出现一个问题，如果是全零端口？这是不可想像的，如果你取了tcp连接的端口，这将是进一步的错误（尽管是同一层级的通信协议，但不意味着等同），这会导致整个工作流程瘫痪或者有特殊的处理。
    - 二是当出现全零端口时，将校验地址模式变更为开放（这意味着将客户端发来的全零地址理解成不需要检验地址）。
    - 让我们分析下实际场景，不难想象这两者的最主要的区别在于受到nat转换的程度，前者因为始终没有放弃检验的步骤，所以在nat转换的过程，是安全的行为；后者因为会出现全零开放的情况，这会导致因为nat转换而变得不安全。
    - 好吧，因为socks5协议诞生的时间比较早，原本我倾向于方案一的解法，但随着我更进一步查阅一些留言（有关socks5的udp代理的资料十分稀少，人们更关注于CONNECT）和与同行进行的辩论，最终我采用了方案二的解法，对这nat转换的行为解释，我认为是socks5产生的时代，nat转换并不是特别常见（尽管现在已经随处可见了）。
    - 让我们进一步思考，目前针对网络环境，并不会有人傻傻的直接使用socks协议在公网上裸奔（如果是，我会笑话他），实际应用的场景基本是本地回环或者是小范围的局域网，放弃nat转换的兼容也是可以理解的（如果想更通用，或许采纳其他的代理协议会更好，这并不是什么难事，并且一个独立的通信协议也会更安全）。
    - 因为产生了开放模式的服务，接收和发送也必定会出现套接字隔离。
  - 结合上诉两点，会得到一个不错的通用方案，并且实际上也有项目采用了，例如v2ray的接入口，回复的server的udp地址是固定的，并没有因为复数的请求而产生独立的套接字。
  - 好吧，不用担心，这只是一些胡言乱语，并不会影响你大部分的使用场景，我能想象到的使用场景基本是一些简单的udp报文转发和quic流代理，如果quic流能达到预期工作状态，我想应该不会有意外的情况。
- 关于CONNECT和BIND我想没有歧义，因为规范文档说明的比较清晰。
- socks协议是一个有趣的通信协议，适合入门的最佳玩意，实现这个库，我能明显感觉到，相比流，报文形式的无状态反而更难处理，也是更需要精力去优化舍取和设计。
## 使用
- ```go
  package main
  
  import (
    "github.com/peakedshout/go-pandorasbox/xnet/proxy/socks"
    "net"
  )
  
  func main() {
    cfg := &socks.ServerConfig{
      VersionSwitch: socks.VersionSwitch{
        SwitchSocksVersion4: true, // socks4/4a
        SwitchSocksVersion5: true, // socks5
      },
      CMDConfig: socks.CMDConfig{
        SwitchCMDCONNECT:          false, // socks4/4a/5 CMDCONNECT
        CMDCONNECTHandler:         nil,   // if nil, use default handler
        SwitchCMDBIND:             false, // socks4/4a/5 BIND
        CMDBINDHandler:            nil,   // if nil, use default handler
        SwitchCMDUDPASSOCIATE:     false, // socks5 UDPASSOCIATE
        CMDCMDUDPASSOCIATEHandler: nil,   // if nil, use default handler
        UDPDataHandler:            nil,   // if nil, use default handler
      },
      Socks5AuthCb: socks.S5AuthCb{
        Socks5AuthNOAUTHPriority:   0,
        Socks5AuthNOAUTH:           nil,
        Socks5AuthGSSAPIPriority:   0,
        Socks5AuthGSSAPI:           nil,
        Socks5AuthPASSWORDPriority: 0,
        Socks5AuthPASSWORD:         nil,
        Socks5AuthIANAPriority:     [125]int8{},
        Socks5AuthIANA:             [125]func(conn net.Conn) net.Conn{},
        Socks5AuthPRIVATEPriority:  [127]int8{},
        Socks5AuthPRIVATE:          [127]func(conn net.Conn) net.Conn{},
      },
      Socks4AuthCb: socks.S4AuthCb{
        Socks4UserIdAuth: nil,
      },
      ConnTimeout: 0,
      DialTimeout: 0,
      BindTimeout: 0,
      UdpTimeout:  0,
    }
    server, _ := socks.NewServer(cfg)
    defer server.Close()
    _ = server.ListenAndServe("tcp", "0.0.0.0:1080")
  }
  ```
- 更多请看test文件。