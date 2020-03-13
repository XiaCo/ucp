[README](https://github.com/XiaCo/ucp/blob/master/README.md) | [中文文档](https://github.com/XiaCo/ucp/blob/master/README.zh.md)

What is ucp?
---
一个用udp写的文件传输工具，适用于远距离传输。
暂时只提供了从有独立ip的终端拉取文件的功能。
以后将开发双端传输或是文件夹同步等等功能。

Installation
---
请从Release页面下载对应可执行文件

Example
---
+ 首先，需要打开远程终端的udp端口，并重启防火墙
```
firewall-cmd --zone=public --add-port=56789/udp
firewall-cmd --reload
```
+ 然后，在远程终端以服务器模式启动文件
```
./ucp -s
```
+ 最后在其它电脑上运行以下，记得替换ip与文件名
```
./ucp -cp "22.22.22.22:56789 /home/test.tar ./test.tar"
```

Usage
---
```
Usage of transfer:
  -cp string
        format: "remoteIP:port filePath savePath"
        example: "22.22.22.22:56789 /home/test.zip ./test.zip"
  -h    help for ucp
  -help
        help for ucp
  -s    run in server
  -saddr string
        file sender listen to binding address (default "0.0.0.0:56789")
  -speed uint
        It is recommended to fill in the minimum bandwidth download / upload speed at both ends
        unit: Kb/s
         (default 1024)
```

Future
---
1. 传输加密，用户验证
2. 双向传输，考虑做成文件夹同步工具
3. 重构