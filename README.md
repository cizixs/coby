# coby: 编写自己的 docker 容器

docker 用到的技术：

- namespaces
- cgroups
- union file system

## 运行

在当前目录创建一个 rootfs，直接使用 busybox 镜像的内容：

```
$ mkdir rootfs/
$
$ docker pull busybox
$ docker run --name coby-busybox -d busybox top -b
$ docker export -o busybox.tar coby-busybox
$ tar -xf busybox.tar -C ./rootfs/
```

编译并运行：

```
$ go build .
$ sudo ./coby run sh
```
