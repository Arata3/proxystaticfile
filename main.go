package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego/httplib"
)

type Conf struct {
	Port     string
	Hosts    []string
	LocalDir string
}

var configFile = "proxystaticfile.toml"
var conf *Conf

var configFile_template = ` # proxystaticfile
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
`

func init() {

	if len(os.Args) > 1 {
		flag.StringVar(&configFile, "c", "proxystaticfile.toml", "extention eg: -c proxystaticfile.toml")
		flag.Parse()
	}

	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		log.Fatal("[conf]", err)
	}
	log.Println(conf)

}

func main() {

	http.HandleFunc("/", findFile)

	log.Println("[server]:start :", conf.Port)
	err := http.ListenAndServe(":"+conf.Port, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func findFile(w http.ResponseWriter, r *http.Request) {

	url := r.URL.String()
	url = strings.Trim(url, "%20")

	log.Println("[geturl]:", url)

	// 本地查询文件
	path := filepath.Clean(conf.LocalDir + url)
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		log.Println("[localfile]:", err)
	} else {
		fileInfo, _ := file.Stat()
		log.Printf("[localfile]:%s is exist.\n", fileInfo.Name())
		fb, _ := ioutil.ReadFile(path)
		w.Write(fb)
		return
	}

	// 远程查询文件
	pf := newProxyFile(url)
	header, data, status := pf.findFile()

	for k, v := range header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(status)
	w.Write(data)

}

// 代理文件请求
type proxyFile struct {
	url   string
	hosts []string
}

// 创建新的
func newProxyFile(url string) proxyFile {
	return proxyFile{
		url:   url,
		hosts: conf.Hosts,
	}
}

// 查询远程文件存在，仅返回200，400的数据
func (this *proxyFile) findFile() (header http.Header, data []byte, status int) {
	// chan findok int

	for _, hostStep := range this.hosts {
		url := "http://" + hostStep + this.url

		req := httplib.Get(url).SetTimeout(time.Second*10, time.Second*10)
		resp, err := req.Response()
		if err != nil {
			log.Println(err)
		}

		status = resp.StatusCode
		header = resp.Header
		log.Printf("[proxy][url]%s [%d]\n", url, status)

		if status == 200 {
			data, _ = req.Bytes()
			return
		}
	}

	status = http.StatusNotFound
	data = []byte("404: File is undefined!")
	return
}
