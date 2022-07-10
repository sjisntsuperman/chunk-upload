package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const tmpFilePath = "./tmp/"
var lock sync.WaitGroup

func main() {
	http.HandleFunc("/merge", chunkHandler)
	err := http.ListenAndServe("0.0.0.0:8001", nil)
	if err != nil {
		log.Fatal("server failed", err)
	}
}

func chunkHandler(w http.ResponseWriter, r *http.Request) {
	setCors(w)
	_, err := mergeChunk(r)
	if err != nil {
		log.Fatal("chunkHanler failed")
	}
}
func setCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}
func mergeChunk(r *http.Request) (int, error) {
	_, err := chunkUpload(r)
	if err != nil {
		return 0, errors.New("upload failed")
	}
	chunkTotal := r.FormValue("chunkTotal")
	fileSize := r.FormValue("fileSize")
	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		return 0, errors.New("file error")
	}
	total, _ := strconv.Atoi(chunkTotal)
	size, _ := strconv.Atoi(fileSize)
	totalLen := 0
	if isFinish(fileHeader.Filename, total, size) {
		filePath := "./" + fileHeader.Filename
		fileBool, err := createFile(filePath)
		if !fileBool {
			return 0, err
		}
		for i := 0; i < total; i++ {
			lock.Add(i)
			go mergeFile(i, fileHeader.Filename, filePath)
		}
		lock.Wait()
	}
	return totalLen, nil
}
func isFinish(fileName string, chunkTotal, fileSize int) bool {
	var chunkSize int64
	for i := 0; i < chunkTotal; i++ {
		iStr := strconv.Itoa(i)
		fi, err := os.Stat(tmpFilePath + fileName + "_" + iStr)
		if err == nil {
			chunkSize += fi.Size()
		}
	}
	return int(chunkSize) == fileSize
}
func mergeFile(idx int, fileName, filePath string) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	defer file.Close()
	if err != nil {
		log.Fatal("file not exists")
	}
	fi, _ := os.Stat(tmpFilePath + fileName + "_0")
	chunkSize := fi.Size()
	file.Seek(chunkSize*int64(idx), 0)
	iSize := strconv.Itoa(idx)
	chunkFilePath := tmpFilePath + fileName + "_" + iSize
	fmt.Printf("chunk path", chunkFilePath)
	chunkFileObj, err := os.Open(chunkFilePath)
	defer chunkFileObj.Close()
	if err != nil {
		log.Fatal("open chunk failed")
	}
	totalLen := 0
	data := make([]byte, 1024, 1024)
	for {
		tal, err := chunkFileObj.Read(data)
		if err == io.EOF {
			chunkFileObj.Close()
			err := os.Remove(chunkFilePath)
			if err != nil {
				fmt.Println("tmp file remove failed", err)
			}
			fmt.Println("copied")
			break
		}
		len, err := file.Write(data[:tal])
		if err != nil {
			log.Fatal("upload failed")
		}
		totalLen += len
	}
	// lock.Done()
	// return totalLen, nil
}
func uploadFile(upfile multipart.File, upSeek int64, file *os.File, fSeek int64) (int, error) {
	fileSize := 0
	upfile.Seek(upSeek, 0)

	file.Seek(fSeek, 0)
	data := make([]byte, 1024)
	for {
		total, err := upfile.Read(data)
		if err == io.EOF {
			break
		}

		len, err := file.Write(data[:total])
		if err != nil {
			return 0, errors.New("uploadFile failed")
		}

		fileSize += len
	}

	return fileSize, nil
}
func createFile(filePath string) (bool, error) {
	fileBool, err := fileExist(filePath)
	if fileBool && err == nil {
		return true, errors.New("file existed")
	} else {
		newFile, err := os.Create(filePath)
		defer newFile.Close()
		if err != nil {
			return false, errors.New("createFile failed")
		}
	}
	return true, nil
}
func fileExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, err
	}
	return false, err
}
func chunkUpload(r *http.Request) (int, error) {
	chunkIndex := r.FormValue("chunkIndex")
	upFile, fileHeader, err := r.FormFile("file")
	if err != nil {
		return 0, errors.New("upload exception")
	}
	filePath := tmpFilePath + fileHeader.Filename + "_" + chunkIndex
	fileBool, err := createFile(filePath)
	if !fileBool {
		return 0, err
	}
	fi, _ := os.Stat(filePath)
	if fi.Size() == fileHeader.Size {
		return 0, errors.New("file existed")
	}
	start := strconv.Itoa(int(fi.Size()))

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	defer file.Close()
	if err != nil {
		return 0, errors.New("openFile failed")
	}
	cnt, _ := strconv.ParseInt(start, 10, 14)
	total, err := uploadFile(upFile, cnt, file, cnt)

	return total, err
}