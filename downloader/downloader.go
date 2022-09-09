package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type Download struct{
	Uuid string
	Url string
	TargetPath string
	TotalSections int
}

func NewDownload(url, targetPath string, totalSections int) Download{
	return Download{
		uuid.NewString(),
		url, 
		targetPath, 
		totalSections,
	}
}

func (d Download) Do() error{
	r, err := d.getNewRequest("HEAD")
	if err != nil{
		return err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil{
		return nil
	}
	fmt.Printf("Got %d", resp.StatusCode)
	if resp.StatusCode > 299{
		return errors.New("cant process response")
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil{
		return err
	}
	fmt.Printf("File size is %d bytes\n", size)

	sections := makeSections(size, d.TotalSections)

	fmt.Println(sections)
	
	err = os.Mkdir(d.Uuid, os.ModePerm)
	if err != nil{
		return err
	}

	var wg sync.WaitGroup
	for i, s := range sections{
		wg.Add(1)
		i := i
		s := s
		go func() {
			defer wg.Done()
			err = d.downloadSection(i, s)
			if err != nil{
				panic(err)
			}
		}()
	}
	wg.Wait()
	err = d.mergeSections(sections)
	if err != nil{
		return err
	}
	err = d.clearTemp()
	if err != nil{
		return err
	}
	return nil
}

func (d Download) mergeSections(sections [][2]int) error{
	f, err := os.OpenFile(d.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil{
		return err
	}
	defer f.Close()
	for i := range sections{
		b, err := ioutil.ReadFile(fmt.Sprintf("%v\\section-%v.tmp", d.Uuid, i))
		if err != nil{
			return err
		}
		_, err = f.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d Download) clearTemp() error {
	return os.RemoveAll(d.Uuid)
}

func (d Download) downloadSection(i int, s [2]int) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s[0], s[1]))
	resp, err := http.DefaultClient.Do(r)
	if err != nil{
		return err
	}
	fmt.Printf("Downloaded %v bytes in for section %v: %v\n", resp.Header.Get("Content-Length"), i, s)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%v\\section-%v.tmp", d.Uuid, i), b, os.ModePerm)
	if err != nil{
		return err
	}
	return nil
}

func (d Download) getNewRequest(method string) (*http.Request, error){
	r, err := http.NewRequest(
		method,
		d.Url,
		nil,
	)
	if err != nil{
		return nil, err
	}
	r.Header.Set("User-Agent", "DownloadManager")
	return r, nil
}

func makeSections(size, totalSections int) [][2]int{
	var sections = make([][2]int, totalSections)
	eachSize := size / totalSections
	for i := range sections{
		if i * eachSize + i >= size - 1{
			sections[i][0] = size - 1
			sections[i][1] = size - 1
			return sections
		}
		if i * eachSize + eachSize + i >= size - 1{
			sections[i][0] = i * eachSize + i
			sections[i][1] = size - 1
			return sections

		}
		sections[i][0] = i * eachSize + i
		sections[i][1] = i * eachSize + eachSize + i
	}
	return sections
}