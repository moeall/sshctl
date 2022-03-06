### 描述

sshctl是一个ssh工具，用于文件传输，和执行远程命令

### 特点

+ 可以屏蔽是使用SSH过程中的交互动作
+ 支持同时对多个远程主机进行操作，减少重复操作
+ 纯Go实现

### 用例

打印帮助信息

> sshctl [-h]

将远程主机上的文件下载到本地

> sshctl get remoteFile localFile -p *** -r 192.168.170.111

将本地文件上传到远程主机

> sshctl put localFile remoteFile -p *** -r 192.168.170.111

直接在远程主机上执行命令

>sshctl sh  hostnamectl -p *** -r 192.168.170.111

