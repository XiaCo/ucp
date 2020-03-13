[README](https://github.com/XiaCo/ucp/blob/master/README.md) | [中文文档](https://github.com/XiaCo/ucp/blob/master/README.zh.md)

What is ucp?
---
A file copy tool using UDP underlying protocol. It is good at transferring files over long distances.
Only the function of downloading files from a machine with a separate IP is provided for the time being. 
In the future, it will provide two-way transfer, folder synchronization and other functions


Installation
---
please download the executable file from the Release page

Example
---
+ first, you should open the port with udp protocol and reload the firewall
```
firewall-cmd --zone=public --add-port=56789/udp
firewall-cmd --reload
```
+ second, execute the transfer in server
```
./ucp -s
```
+ third, execute the transfer in client.It copies file /home/test.tar on remote to the current directory
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
1. Transmission encryption. User Authentication
2. Two-way transfer, consider making a folder synchronization tool
3. Refactoring