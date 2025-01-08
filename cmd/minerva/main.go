package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	uniqueIPs := make(map[string]bool)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		ips := ipRegex.FindAllString(line, -1)
		for _, ip := range ips {
			uniqueIPs[ip] = true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	for ip := range uniqueIPs {
		fmt.Println(ip)
	}
}
