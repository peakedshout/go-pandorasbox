# socks
#### (EN/[CN](./README_CN.md))
## Explanation
- This is a pure golang implementation of socks4/4a/5. It refers to the specification documents, integrates abstract interfaces and supports more detailed configurations. Please note that the operable space of this library is very huge. If you don’t understand what you are doing , then it is best to do nothing and use the default method.
- As for why this thing exists? This is very simple, because the socks currently implemented in golang cannot completely satisfy me, and this socks is not very difficult.
- This implementation has been tested by multiple socks clients and servers, and there are no errors that affect use.
- Note that this implementation is based on a brief specification document. Because the document does not carefully guide the details, it is normal for disputes to exist.
## Theory
- Before reading the principles, I think there is nothing more straightforward than reading the specification document. Here are the signposts:
  - socks4/4a： not found rfc (or you tell me)
  - socks5：rfc1929  rfc1928
- socks4 only has CONNECT and BIND methods, and it is a tcp stream. The user authentication method is userid. It may be more direct and universal to understand this userid as sessionid (?)
- socks4a is a domain name patch for socks4, because when socks4 was designed, it did not consider the issue of ipv6 and domain name, and socks4a only solved the domain name issue. Compared to that time, ipv6 had not been taken into consideration (?)
- socks5 Compared with the first two, socks5 has the UDPASSOCIATE method, which is a method that supports udp proxy forwarding, and socks5 supports ipv6 and better supports domain names.
- Note that because the description of the specification document is very brief, there is ambiguity in UDPASSOCIATE of socks5. Further explanation is given below:
  - Does the UDP socket created by this server participate in direct message reception? (That is, is it the receiver and the sender?)
    - Regarding this point, the description in the specification document is very vague. Regarding the security of socks, socket isolation for receiving and sending seems to be more in line with the proxy concept of socks (in fact, there are implementations that do this)
    - However, the isolation of receiving and sending will generate more maintenance overhead. This depends on personal choice. This implementation uses an isolation solution.
  - Is the client not sending the client's expected address? (i.e. all-zero address, invalid address, here refers to all-zero ip and all-zero port)
    - Regarding this point, the description in the specification document is very vague. There are only two solutions implemented by developers:
    - The first is to take the sender's IP as the client's address for service. This is obviously the simplest, but there will be a problem, what if it is an all-zero port? This is unimaginable. If you take the port of the TCP connection, this will be a further error (although it is the same level of communication protocol, it does not mean the same), which will cause the entire workflow to be paralyzed or have special processing.
    - The second is to open the verification address mode change when an all-zero port appears (this means that the all-zero address sent by the client is understood as an address that does not need to be verified).
    - Let us analyze the actual scenario. It is not difficult to imagine that the main difference between the two is the degree of NAT conversion. The former does not give up the verification step, so it is a safe behavior during the NAT conversion process; the latter because it will There is a situation where all zeros are open, which will cause it to become unsafe due to NAT conversion.
    - Well, because the socks5 protocol was born relatively early, I originally preferred solution 1, but as I further checked some comments (information about socks5’s udp proxy is very scarce, people are more focused on CONNECT) and colleagues After a debate, I finally adopted solution 2. To explain the behavior of NAT conversion, I think it was the era when socks5 was produced, and NAT conversion was not particularly common (although it can be seen everywhere now).
    - Let us think further. Regarding the current network environment, no one will use the socks protocol directly to streak on the public network (if so, I would laugh at him). The actual application scenarios are basically local loopback or small-scale LAN. It is understandable to give up the compatibility of NAT conversion (if you want to be more versatile, it may be better to adopt other proxy protocols. This is not difficult, and an independent communication protocol will be more secure).
    - Because of the creation of open-mode services, socket isolation must also occur for receiving and sending.
  - Combining the two points above, you will get a good general solution, and some projects have actually adopted it. For example, for the access port of v2ray, the UDP address of the reply server is fixed, and no independent sockets are generated due to multiple requests.
  - Well, don’t worry, this is just some nonsense and will not affect most of your usage scenarios. The usage scenarios I can imagine are basically some simple udp message forwarding and quic flow proxy. If quic flow can achieve As expected, I don't think there will be any surprises.
- I think there is no ambiguity about CONNECT and BIND because the specification document explains it clearly.
- The socks protocol is an interesting communication protocol, the best thing for getting started. When implementing this library, I can clearly feel that compared to streams, the statelessness in the form of messages is more difficult to process, and it also requires more energy to optimize rounding and design.
## Use
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
- Please see the test file for more information.