### 描述

sshctl是一个用于文件传输、执行远程脚本的无交互的ssh工具

### 特点

+ 可以屏蔽是使用SSH过程中的交互动作，进而可实现完全自动化
+ 支持同时对多个远程主机进行操作，减少重复操作
+ 多任务异步执行，高效快速
+ 纯Go实现，无依赖，体积小

### 用例

打印帮助信息

```shell
sshctl [-h]
```

将远程主机上的文件下载到本地

```shell
sshctl get remoteFile localFile -p *** -r 192.168.170.111
```

将本地文件上传到远程主机

```shell
sshctl put localFile remoteFile -p *** -r 192.168.170.111
```

在远程主机执行命令

```shell
sshctl sh -c hostnamectl -p *** -r 192.168.170.111
```

在远程主机上执行本地脚本

```shell
sshctl sh -f  script.sh -p *** -r 192.168.170.111
```

在远程主机上执行本地从本地pipe接受到的脚本

```shell
echo scripth.sh | sshctl -p *** -r 192.168.170.111 sh -f  -
```



