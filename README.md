# proxy static file
> 静态中心代理服
> 临时工具

**使用场景**

主要用于分布式服务器的相对位置文件的轮训扫描，与nginx配合使用。
当使用多台服务器的负载均衡处理，上传无中心处理的时候, A 文件在 S1 服务上传了文件， B文件在 S2 服务器需要访问，
这个时候由代理服务来轮训（现在为顺序轮训）节点服务的资源，然后返回资源并备份到当前服务中心。

用户访问的页面请求的静态资源走CDN服务器（可不用），CDN找不到文件，CDN会从源服务器（资源中心）获取文件，
nginx 检查本地文件是否存在，不存在则使用代理轮训已经配置的节点服务器，轮训查找后则将文件返回并在本地进行
备份存储。

寻寻找文件的优先级为 "CDN -> 源主机 -> 轮询其他主机的位置"

> 关于轮训的方式，当前的仅做了顺序轮训，后面利用 golang 的特性 使用 `sync.WaitGroup` 计数或者
> 使用并行通信 `chan` 来并发加速查询，加快查询的速度。

* **port** 为服务运行端口，之后交给nginx代理使用
* **hosts** 为节点服务器的地址，不带有 `http://` 的相对的 `URL` 路径
* **localDir** 为存储文件的本地地址，需要配置，因为要检查文件
* **WriteHere** 是否写入

```
# proxystaticfile
# 使用  ./proxystaticfile -c ./proxystaticfile.toml
# 默认挂载当前的的conf文件，配置文件使用 toml 格式输出
# 【情景】
#  用于代理均衡负载分布服务器的的静态文件输出，默认超时时间为30s

# 运行端口
port = "8808"
# 远程地址，在局域网中使用局域网地址；注意文件的相对路径
hosts = ["127.0.0.1","www.xxx.net/img"]
# 当前服务器端的文件所在地址
localDir = "/Users/user/Sites/img/"
# 代理的内容自动写入代理服务器
WriteHere = false
```

nginx 配置：

```nginx

location /(css|js|fonts|img)/ {
    access_log off;
    expires 1d;

    root "/path/to/app_b/static"
    try_files $uri @backend
}

location /uploads/ {
    try_files /_not_exists_ @backend;
}


location @backend {
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header Host            $http_host;

    proxy_pass http://127.0.0.1:8081;
}

```


### 关于其他的方案

nginx 代理方案：
    利用 nginx 的轮训查找代理功能代理其他的服务器查询
ssh 硬盘挂载：
    利用 远程挂载硬盘来拼接路径
监控推送：
    监控变动文件，并传送到

问题： 文件实时性，节点的数量变动后改动大，文件的定位已经安全性缺乏。
