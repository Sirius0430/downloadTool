package main

import (
	"fmt"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	//terminal输入go get -u -v github.com/vbauerster/mpb 下载进度条
	time1 := time.Now()
	wg := &sync.WaitGroup{}
	p := mpb.New(mpb.WithWaitGroup(wg))
	pool := make(chan int, 4)          //这里修改协程数
	downloadList := map[string]string{ //这里修改下载列表，格式为 下载后的文件名：下载路径
		"zotero1.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero2.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero3.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero4.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero5.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero6.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero7.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		"zotero8.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero9.exe":  "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero10.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero11.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero12.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero13.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero14.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero15.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
		//"zotero16.exe": "https://www.zotero.org/download/client/dl?channel=release&platform=win32&version=6.0.20",
	}
	for name, url := range downloadList {
		wg.Add(1)
		pool <- 1
		go download(name, url, wg, p, &pool)
	}

	wg.Wait()
	timeUsed := time.Since(time1)
	fmt.Println(timeUsed)
}

func download(filename string, url string, wg *sync.WaitGroup, progress *mpb.Progress, pool *chan int) error {
	//创建临时文件
	tmpName := strings.Split(filename, ".")[0] + ".tmp"
	tmpFile, err := os.Create(tmpName)
	if err != nil {
		tmpFile.Close()
		return err
	}

	//发起请求
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
		//return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		//return err
		return err
	}
	defer res.Body.Close()

	//进度条配置
	fileSize, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	bar := progress.AddBar(
		int64(fileSize),
		// 进度条前的修饰
		mpb.PrependDecorators(
			decor.Name(filename+": "),
			decor.CountersKibiByte("% .2f / % .2f"), // 已下载数量
			decor.Percentage(decor.WCSyncSpace),     // 进度百分比
		),
		// 进度条后的修饰
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)

	reader := bar.ProxyReader(res.Body)
	defer reader.Close()

	//将下载的数据存储到临时文件中
	_, err = io.Copy(tmpFile, reader)
	tmpFile.Close()
	if err != nil {
		log.Fatalln(err)
		return err
	}

	//将tmp文件重命名
	err = os.Rename(tmpName, filename)

	//删除意外退出时残留的tmp文件，并记录残留文件名
	_, err = os.Stat(tmpName)
	if err == nil {
		os.Remove(tmpName)
	}
	<-*pool
	defer wg.Done()
	return nil
}
