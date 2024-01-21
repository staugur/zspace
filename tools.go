package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

var (
	title_pat   = regexp.MustCompile(`<title>(.*?)</title>`)
	emptyLineRe = regexp.MustCompile(`^\s*$`)
)

func IsEmptyLine(str string) bool {
	return emptyLineRe.MatchString(str)
}

func TrimSpaceWS(str string) string {
	return strings.TrimRight(str, " \n\t")
}

func check_zspace_title(ip string, port uint16) (title string) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d", ip, port))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	matches := title_pat.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		title = matches[1]
	}
	return
}

func update_zspace_hostip() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Equivalent to `nmap -n --open -p 5055 192.168.0.100-101`,
	// with a 5-minute timeout.
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets(scan_net),
		nmap.WithPorts(scan_port),
		nmap.WithOpenOnly(),
	)
	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	result, warnings, err := scanner.Run()
	if len(*warnings) > 0 {
		log.Printf("run finished with warnings: %s\n", *warnings) // Warnings are non-critical errors from nmap.
	}
	if err != nil {
		log.Fatalf("unable to run nmap scan: %v", err)
	}

	// Use the results to print an example output
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		fmt.Printf("Host %q:\n", host.Addresses[0])
		for _, port := range host.Ports {
			fmt.Printf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)

			// 可能是极空间设备，进一步测试
			if port.State.String() == "open" {
				address := host.Addresses[0].String()
				title := check_zspace_title(address, port.ID)
				if strings.Contains(title, "极空间") {
					// 更改hosts
					hf := newHostfile("/etc/hosts")
					if ip, has := hf.HasDomain(dname); has == true {
						log.Printf("found exists with ip: %s, delete it", ip)
						hf.DeleteDomain(dname)
					}
					hf.AppendHost(dname, address)
					log.Printf("write device hostname: %s %s", address, dname)
				}
			}
		}
	}

	fmt.Printf("Nmap done: %d hosts up scanned in %.2f seconds\n", len(result.Hosts), result.Stats.Finished.Elapsed)
}
