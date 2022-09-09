package main

import (
	"fmt"
	"time"
	"download_manager/downloader"
)


func main(){
	startTime := time.Now()
	d := downloader.NewDownload(
		"https://upload.wikimedia.org/wikipedia/commons/9/9a/Gull_portrait_ca_usa.jpg",
		"file.jpeg",
		10,
	)
	err := d.Do()
	if err != nil{
		fmt.Printf("Error while downloading file(%s)\n%v\n", d.Url, err)
	}
	fmt.Printf("Download finished in %v seconds\n", time.Since(startTime).Seconds())
}

