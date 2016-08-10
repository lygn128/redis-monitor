package main

import (
	//"net/http"
	//"strings"
	//"os"
	//"log"
	//"io/ioutil"
	//"io/ioutil"
	//"log"
	//"strings"
	"strings"
	"fmt"
	"gopkg.in/redis.v4"
	"time"
	"strconv"
	//"sync"
	"github.com/bluebreezecf/opentsdb-goclient/client"
	"github.com/bluebreezecf/opentsdb-goclient/config"
	"net/http"
	"sync"
	"os"
	"io/ioutil"
	"flag"
)


type redis_info struct {
	rss 	int64
	mem 	int64
	ops 	int64
	timestamp int64
}


const (
loginweb string = "<html><body><h1>add redis server</h1><form action=\"\" method=\"post\">ipaddr:<input type=\"text\" name=\"username\"><br>port     :<input type=\"password\" name=\"pwd\"><br><input type=\"submit\" value=\"add\"><br><a href=\"/regist\">注册</a></form></body></html>"
)
//var stats map[string] []*redis_info

type infomanger struct {
	stats    map[string] string
        rwLocker *sync.RWMutex
}

func (mgr*infomanger) insertNewAddr(addr string) {
	mgr.rwLocker.Lock()
	defer mgr.rwLocker.Unlock()
	onestat := mgr.stats[addr]
	if onestat == "" {
		mgr.stats[addr] = "true"
		return
	} else {
		fmt.Println("insert")
	}
}

func (mgr*infomanger) listAllInstance()string {
	mgr.rwLocker.Lock()
	defer mgr.rwLocker.Unlock()
	xxx := ""
	for key,_ := range mgr.stats {
		xxx = fmt.Sprintln("%s%s\r\n",xxx,key)
	}
	return  xxx
}


func (mgr*infomanger) LoadFromDisk(file string)(err error) {
	mgr.rwLocker.Lock()
	defer mgr.rwLocker.Unlock()
	addrs, eror := ioutil.ReadFile(file)
	if eror != nil {
		return eror
	}

	infoarry := strings.Split(string(addrs),"\n")
	for _,value := range infoarry {
		fmt.Println("arrry",value)
		mgr.stats[value] = "true"
	}
	return
}






func NewinfoManger()(mgr *infomanger) {
	mgr          = new(infomanger)
	mgr.stats    = make(map[string]string)
	mgr.rwLocker = new(sync.RWMutex)
	return mgr
}


var mgr *infomanger = NewinfoManger()



func InsertInfo(addr string) {
     mgr.insertNewAddr(addr)
}





func listAddr()(str string) {
	return mgr.listAllInstance()
}



func staticResource(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	fmt.Println("path:", path)
	if path == "/add" {
		fin, err  := os.Open("/home/yfliuyonggang/redis_monitor/sources/add.js")
		fmt.Println("err",err)
		fd, _ := ioutil.ReadAll(fin)
		w.Write(fd)
		return
	}
	if path == "/" {

               w.Write(([]byte(listAddr())))
		return
	}

}


//func staticResource2(w http.ResponseWriter, r *http.Request) {
//	path := r.URL.Path
//	fmt.Println("path:", path)
//	if path == "/" {
//		datapoints := getMemoryInfo("172.23.5.45:6379")
//		if datapoints == "" {
//			w.Write([]byte("wushuju"))
//			return
//		}
//		xxxx  = strings.Replace(xxxx,"DATA",datapoints, 1)
//		fmt.Println("xxxx", xxxx)
//		w.Write([]byte(xxxx))
//		return
//	}
//
//
//	request_type := path[strings.LastIndex(path, "."):]
//	switch request_type {
//	case ".css":
//		w.Header().Set("content-type", "text/css")
//	case ".js":
//		w.Header().Set("content-type", "text/javascript")
//	default:
//	}
//
//	fin, err := os.Open("/home/yfliuyonggang/redis_monitor/sources/jscharts.js")
//	defer fin.Close()
//	if err != nil {
//		log.Fatal("static resource:", err)
//	}
//	fd, _ := ioutil.ReadAll(fin)
//	w.Write(fd)
//}
var dataPointsBuff  chan client.DataPoint = make(chan client.DataPoint,2000)

func monitor(redisaddr string) {
        redisclient := redis.NewClient(&redis.Options{
		Addr:     redisaddr,
		Password: "", // no password set
		DB:       0,})
	redisname := strings.Replace(redisaddr, ":", "-", 1)

	for {
		result := redisclient.Info()

		statinfo := new(redis_info)
		statinfo.rss = 0
		infoarry := strings.Split(result.Val(),"\r\n")
                itemnum := 0
		for index,value := range infoarry {
			fmt.Println(index, value)
			suarry := strings.Split(value, ":")
			if len(suarry) < 2 {
				continue
			}
			switch suarry[0] {
			case "used_memory":{
				xxx,errr  := strconv.ParseInt(suarry[1], 10, 64)
				fmt.Println("usedmemory", xxx, suarry[1], errr)
				statinfo.mem = xxx/1024/1024
				itemnum++
			}

			case "used_memory_rss":{
				xxx,_      := strconv.ParseInt(suarry[1], 10, 64)
				statinfo.rss = xxx/1024/1024
				fmt.Println("rsssy", xxx)
				itemnum++
			}

			case "instantaneous_ops_per_sec":{
				xxx,_      := strconv.ParseInt(suarry[1], 10, 64)
				fmt.Println("ops", xxx)
				statinfo.ops = xxx
				fmt.Println("xfafd",statinfo.ops)
				itemnum++
			}

			default:
			}
			if itemnum == 3 {
				statinfo.timestamp = time.Now().Unix()
				break
			}

		}
		rss := client.DataPoint{
			Metric:    redisname,
			Timestamp: statinfo.timestamp ,
			Value:     statinfo.rss,
			Tags:map[string]string{
				"id": "rss",
			},
		}

		mem := client.DataPoint{
			Metric:redisname,
			Timestamp:  statinfo.timestamp,
			Value:statinfo.mem,
			Tags:map[string]string{
				"id": "mem",
			},
		}

		qps := client.DataPoint{
			Metric:redisname,
			Timestamp:statinfo.timestamp,
			Value:statinfo.ops,
			Tags:map[string]string{
				"id": "ops",
			},
		}


		dataPointsBuff <- rss
		dataPointsBuff <- mem
		dataPointsBuff <- qps


		//m3 := map[string]string{
		//	"a": "aa",
		//	"b": "bb",
		//}
		//response, err := tsdbclient.Put([]client.DataPoint{dataPoint} , "details")
		//fmt.Println("put result", response, err)
		//InsertInfo(meminfo, "172.23.5.45:6379")
		time.Sleep(time.Second * 60)
		fmt.Println("monitor")

	}
}

func InintServer() {

}

func main() {
        dataPath := flag.String("f","./addrs","file name ")
	flag.Parse()
        err := mgr.LoadFromDisk(*dataPath)
	fmt.Println("load err",err)
	http.HandleFunc("/",staticResource)
	go http.ListenAndServe("0.0.0.0:7777",nil)
        for key,_ := range mgr.stats {
		fmt.Println("value",key)
		go monitor(key)
	}

	tsdbConfg := config.OpenTSDBConfig{OpentsdbHost:"172.28.247.127:4242"}
	tsdbclient,_:= client.NewClient(tsdbConfg)
	for {
		point := <-dataPointsBuff
		_, err := tsdbclient.Put([]client.DataPoint{point} , "details")
		if err!= nil {
			fmt.Println(err)
		}
		fmt.Println("send to tsdb")
	}

}
