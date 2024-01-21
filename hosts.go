package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"tcw.im/gtc"
)

type hostname struct {
	Comment string
	Domain  string
	IP      string
}

func newHostname(comment string, domain string, ip string) *hostname {
	return &hostname{comment, domain, ip}
}

func (h *hostname) toString() string {
	return h.IP + " " + h.Domain + "\n"
}

func appendToFile(filePath string, hostname *hostname) {
	fp, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed opening file %s : %s", filePath, err)
		return
	}
	defer fp.Close()

	_, err = fp.WriteString(hostname.toString())
	if err != nil {
		log.Printf("failed append string: %s: %s", filePath, err)
		return
	}
}

type HostFile struct {
	Path string
}

func newHostfile(path string) *HostFile {
	return &HostFile{Path: path}
}

func (h *HostFile) ParseHostFile(path string) (*linkedhashmap.Map, error) {
	if !gtc.IsFile(path) {
		log.Printf("path %s is not exists", path)
		return nil, errors.New("path %s is not exists")
	}

	fp, fpErr := os.Open(path)
	if fpErr != nil {
		log.Printf("open file '%s' failed", path)
		return nil, fmt.Errorf("open file '%s' failed ", path)
	}
	defer fp.Close()

	br := bufio.NewReader(fp)
	lm := linkedhashmap.New()
	curComment := ""
	for {
		str, rErr := br.ReadString('\n')
		if rErr == io.EOF {
			break
		}
		if len(str) == 0 || str == "\r\n" || IsEmptyLine(str) {
			continue
		}

		if str[0] == '#' {
			// 处理注释
			curComment += str
			continue
		}
		tmpHostnameArr := strings.Fields(str)
		curDomain := strings.Join(tmpHostnameArr[1:], " ")
		//if !iputils.CheckDomain(curDomain) {
		// return lm, errors.New(" file contain error domain" + curDomain)
		//}
		curIP := TrimSpaceWS(tmpHostnameArr[0])

		checkIP := net.ParseIP(curIP)
		if checkIP == nil {
			continue
		}
		tmpHostname := newHostname(curComment, curDomain, curIP)
		lm.Put(tmpHostname.Domain, tmpHostname)
		curComment = ""
	}

	return lm, nil
}

func (h *HostFile) AppendHost(domain string, ip string) {
	if domain == "" || ip == "" {
		return
	}

	hostname := newHostname("", domain, ip)
	appendToFile(h.Path, hostname)
}

func (h *HostFile) writeToFile(hostnameMap *linkedhashmap.Map, path string) {
	if !gtc.IsFile(path) {
		log.Printf("path %s is not exists", path)
		return
	}

	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("open file '%s' failed: %v", path, err)
		return
	}
	defer fp.Close()

	hostnameMap.Each(func(key interface{}, value interface{}) {
		if v, ok := value.(*hostname); ok {
			_, writeErr := fp.WriteString(v.toString())
			if writeErr != nil {
				log.Println(writeErr)
				return
			}
		}
	})
}

func (h *HostFile) DeleteDomain(domain string) {
	if domain == "" {
		return
	}

	currHostsMap, parseErr := h.ParseHostFile(h.Path)
	if parseErr != nil {
		log.Printf("parse file failed" + parseErr.Error())
		return
	}
	_, found := currHostsMap.Get(domain)
	if currHostsMap == nil || !found {
		return
	}
	currHostsMap.Remove(domain)
	h.writeToFile(currHostsMap, h.Path)
}

func (h *HostFile) HasDomain(domain string) (ip string, has bool) {
	if domain == "" {
		return
	}

	currHostsMap, parseErr := h.ParseHostFile(h.Path)
	if parseErr != nil {
		log.Printf("parse file failed" + parseErr.Error())
		return
	}
	value, found := currHostsMap.Get(domain)
	if currHostsMap == nil || !found {
		return
	}
	if v, ok := value.(*hostname); ok {
		return v.IP, true
	}
	return
}

func (h *HostFile) ListCurrentHosts() {
	currHostsMap, parseErr := h.ParseHostFile(h.Path)
	if parseErr != nil {
		log.Printf("parse file failed" + parseErr.Error())
		return
	}
	if currHostsMap == nil {
		return
	}
	currHostsMap.Each(func(key interface{}, value interface{}) {
		if v, ok := value.(*hostname); ok {
			fmt.Print(v.toString())
		}
	})
}
