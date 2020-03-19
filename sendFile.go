package GoTools

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

/*
Http send file
f		:File pointer
buf_n	:File buffer size. Default:512
 */
func sendFile(writer http.ResponseWriter, request *http.Request, f *os.File,buf_n int) {
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		log.Println("sendFile1", err.Error())
		http.NotFound(writer, request)
		return
	}
	writer.Header().Add("Accept-Ranges", "bytes")
	writer.Header().Add("Content-Disposition", "attachment; filename="+info.Name())

	etag := sha1.New()
	etag.Write([]byte(strconv.FormatInt(info.ModTime().UnixNano(), 10)))
	writer.Header().Add("ETag", fmt.Sprintf("%x", etag.Sum(nil)))
	var start, end int64
	IfRange := request.Header.Get("If-Range")
	//fmt.Println(request.Header,"\n")
	if r := request.Header.Get("Range"); r != "" && (IfRange == fmt.Sprintf("%x", etag.Sum(nil)) || IfRange == "") {
		if strings.Contains(r, "bytes=") && strings.Contains(r, "-") {

			fmt.Sscanf(r, "bytes=%d-%d", &start, &end)
			if end == 0 {
				end = info.Size() - 1
			}
			if start > end || start < 0 || end < 0 || end >= info.Size() {
				writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				log.Println("sendFile2 start:", start, "end:", end, "size:", info.Size())
				return
			}
			writer.Header().Add("Content-Length", strconv.FormatInt(end-start+1, 10))
			writer.Header().Add("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, end, info.Size()))
			writer.WriteHeader(http.StatusPartialContent)
		} else {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		writer.Header().Add("Content-Length", strconv.FormatInt(info.Size(), 10))
		start = 0
		end = info.Size() - 1
	}
	_, err = f.Seek(start, 0)
	if err != nil {
		log.Println("sendFile3", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	buf := make([]byte, buf_n)
	for {
		if end-start+1 < int64(buf_n) {
			buf_n = int(end - start + 1)
		}
		_, err := f.Read(buf[:buf_n])
		if err != nil {
			log.Println("1:", err)
			if err != io.EOF {
				log.Println("error:", err)
			}
			return
		}
		err = nil
		_, err = writer.Write(buf[:buf_n])
		if err != nil {
			//log.Println(err, start, end, info.Size(), n)
			return
		}
		start += int64(buf_n)
		if start >= end+1 {
			return
		}
	}

}