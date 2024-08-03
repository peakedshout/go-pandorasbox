# go-pandorasbox
## Integrated Tool Library [Pandora's Box]
#### Concept: Integrated, efficient, abstract, diverse, reliable
### Language
#### (EN/[CN](./README_CN.md))
## What is it? How does it work?
- This is an integrated tool library, currently mainly partial network tools, such as communication protocols, encryption coding, RPC frameworks, etc., combined with my thinking in Golang development, independently summarized and encapsulated development tool library.
- Because this is an integrated tool library, you can check what functions are available through [Index](#index). It is expected that each key functional module will have a description file later.
- Because this tool library originally served the anchor project, the functionality does not fully cover any needs.
- At the same time, this is an open image that is released after certain sorting and screening. Therefore, if there are missing dependencies, it may be that the source project has not been synchronized. Please wait patiently for a while. (The source project is full of testing and debugging stuff, I guess no one likes this)
- As for why I opened up this image, firstly, [anchorage](https://github.com/peakedshout/go-anchorage) needs open source distribution and this tool library needs to be made public; secondly, this tool library has helped me a lot in development, and I think good things are worth sharing.
## Can this thing be put into production? Can safety be guaranteed?
- Unfortunately, this tool library is only maintained by me, and review and debugging are performed by me. My personal abilities are limited, and I cannot guarantee that this tool library can support your work 100%.
- If you want it to go into production, I expect it to, but I suggest you review the references to reduce the chance of problems.
## Other
- If you find this tool library useful, please give me a ✨ to let more people know about it.
- If you have related questions, you can submit issues and I will deal with it when I am free.
## Guess you want to see
1. [httpporxy](./xnet/proxy/httpproxy)
   - This is a simple proxy implementation of http. It uses the net/http package internally, so it only supports http1 type protocols.
2. [socks](./xnet/proxy/socks)
   - This is a pure golang implementation of socks4/4a/5. It refers to the specification document, integrates abstract interfaces and supports more detailed configuration.
3. [xrpc](./xrpc)
   - xrpc is a custom-implemented rpc communication framework.
## Index
1. [_build](./_build)
    - This is a script for simple cross-compilation. It is usually copied to your own project and then run to compile the project.
2. [ccw](./ccw)
    - Here are some commonly used asynchronous control tool packages.
    - Among them, ctxtools encapsulates and expands context. The functionality of this expansion exceeds my expectations. Now my project must reference it in asynchronous operations.
3. [control](./control)
    - This is an encapsulation of the underlying socket communication.
4. [logger](./logger)
    - This is a self-developed log library for printing and debugging; it has the functions of log classification, log distribution, and log delivery.
5. [pcrypto](./pcrypto)
    - The encapsulated encryption library abstracts a unified interface, making polymorphic encryption easier to call.
6. [protocol](./protocol)
    - The encapsulated coding protocol library implements a unified interface, and the coding layer is indispensable in network communication. If you are interested, you can take a look at the relevant implementations.
7. [tool](./tool)
    - Various messy tool libraries basically only implement some simple functions, and are not so large that they can be singled out. For simple classification, they are stacked in tools.
    - If you are interested, you can take a look at the directory at this level.
8. [uerror](./uerror)
    - The error library that implements the only error code solves the difficulty of cross-project and locating errors, and implements the function of related errors. However, I am not very satisfied with this library, and some libraries have references.
    - Perhaps a simpler implementation should be used to replace it in the future.
9. [xmsg](./xmsg)
    - xmsg is a custom implementation of streaming communication protocol, custom message structure, and integrated records of delay, rate and life cycle.
    - xmsg is session-oriented.
    - This application scenario focuses on low-level connection processing. Implementing this set of methods has taught me a lot about network knowledge.
10. [xnet](./xnet)
    - The integrated network tool library includes tools for implementing network interfaces, network agents, etc. It is a convenient scaffolding for me when developing network logic.
    - If you are interested, you can take a look at the directory at this level.
11. [xrpc](./xrpc)
    - xrpc is a custom-implemented rpc communication framework. Based on the xmsg library, it implements the communication of rpc, stream, send stream, recv stream, and rrpc.
    - The focus of xrpc is a stream-oriented working method, in which the relationship between sessions and streams is a 1:n relationship.
    - This application scenario refers to the workflow of grpc and combines it with the rpc communication framework completed by my own thinking. Similarly, implementing this set has taught me a lot about thinking about network knowledge.
## TODO
1. [ ] Greater reliability (personal tool library, cannot guarantee 100% suitability for all projects)
2. [ ] Stronger underlying implementation (sorry, algorithm classes are my weakness)
3. [ ] More expansion (involving my knowledge blind spots)
4. [ ] More documentation (but I'm lazy, that's all I can hope for)