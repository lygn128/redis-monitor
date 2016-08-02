package main

import (
	"net/http"
	//"strings"
	//"os"
	//"log"
	//"io/ioutil"
	"io/ioutil"
	"log"
	"os"
	//"strings"
	"strings"
	"fmt"
	"gopkg.in/redis.v4"
	"time"
	"strconv"
)


type redis_info struct {
	rss int64
	mem int64
	ops int
}

var stats map[string] []*redis_info

func staticResource(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	fmt.Println("path:", path)
	if path == "/" {
		fin, err := os.Open("/home/yfliuyonggang/redis_monitor/examples/line-charts/example-1/index.html")
		defer fin.Close()
		if err != nil {
			log.Fatal("static resource:", err)
		}
		fd, _ := ioutil.ReadAll(fin)
		w.Write(fd)
		return
	}


	request_type := path[strings.LastIndex(path, "."):]
	switch request_type {
	case ".css":
		w.Header().Set("content-type", "text/css")
	case ".js":
		w.Header().Set("content-type", "text/javascript")
	default:
	}

	fin, err := os.Open("/home/yfliuyonggang/redis_monitor/sources/jscharts.js")
	defer fin.Close()
	if err != nil {
		log.Fatal("static resource:", err)
	}
	fd, _ := ioutil.ReadAll(fin)
	w.Write(fd)
}


func monitor() {
	stats = make(map [string][]*redis_info)
        client := redis.NewClient(&redis.Options{
		Addr:     "172.23.5.46:6379",
		Password: "", // no password set
		DB:       0,})
	for {
		result := client.Info("memory")
		fmt.Println(result.Val())

		meminfo    := new(redis_info)
		meminfo.rss = 0
		infoarry := strings.Split(result.Val(),"\n")
		fmt.Println(len(infoarry))

		for index,value := range infoarry {
			fmt.Println(index, value)
			suarry := strings.Split(value, ":")
			if len(suarry) < 2 {
				continue
			}
			switch suarry[0] {
			case "used_memor":{
				xxx,_      := strconv.ParseInt(suarry[1],10,8)
				meminfo.mem = xxx
			}

			case "used_memory_rss":
				xxx,_      := strconv.ParseInt(suarry[1],10,8)
				meminfo.rss = xxx
			default:
			}
			onestat := stats["172.23.5.46:637"]
			if onestat == nil {
				stats["172.23.5.46:637"] = []*redis_info{meminfo}
				continue
			}

			onestat = append(onestat,meminfo)
			stats["172.23.5.46:637"] = onestat

		}
		time.Sleep(time.Second * 60)

	}
}


func main() {
	http.HandleFunc("/",staticResource)
	go http.ListenAndServe("0.0.0.0:7777",nil)
	monitor()

}
