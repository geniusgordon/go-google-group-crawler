package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func MkdirAll(group string) {
	fmt.Printf(":: mkdir %s/\n", group)
	os.MkdirAll(fmt.Sprintf("%s/threads", group), 0700)
	os.MkdirAll(fmt.Sprintf("%s/msgs", group), 0700)
	os.MkdirAll(fmt.Sprintf("%s/mbox", group), 0700)
	os.MkdirAll(fmt.Sprintf("%s/mbox/cur", group), 0700)
	os.MkdirAll(fmt.Sprintf("%s/mbox/new", group), 0700)
	os.MkdirAll(fmt.Sprintf("%s/mbox/tmp", group), 0700)
}

func DumpLinksFromUrl(t string, group string, url string, output_filename string) int {
	r_total, _ := regexp.Compile("<i>.*?([0-9]+).*?([0-9]+).*?([0-9]+).*?</i>")
	r_url, _ := regexp.Compile(fmt.Sprintf("\"(https?://.*?/d/%s/%s.*?)\"", t, group))

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	output_file, err := os.Create(output_filename)
	if err != nil {
		log.Fatal(err)
	}

	total := 0

	for scanner.Scan() {
		text := scanner.Text()
		if r_total.MatchString(text) {
			match := r_total.FindStringSubmatch(text)
			a, _ := strconv.Atoi(match[1])
			b, _ := strconv.Atoi(match[2])
			c, _ := strconv.Atoi(match[3])
			t := []int{a, b, c}
			sort.Ints(t)
			total = t[2]
		} else if r_url.MatchString(text) {
			match_url := r_url.FindStringSubmatch(text)[1]
			output_file.WriteString(match_url)
			output_file.WriteString("\n")
		}
	}
	return total
}

func DownloadPageWorker(id int, t string, group string, url string, output_prefix string, jobs <-chan [2]int, results chan<- int) {
	for r := range jobs {
		fmt.Printf(":: Downloading %s from %s[%d-%d] by worker[%d]\n", t, group, r[0], r[1], id)
		url_with_range := fmt.Sprintf("%s[%d-%d]", url, r[0], r[1])
		output_filename := fmt.Sprintf("%s.%d.%d", output_prefix, r[0], r[1])
		DumpLinksFromUrl(t, group, url_with_range, output_filename)
	}
	results <- id
}

func DownloadPages(t string, group string, url string, output_prefix string, workers int) {
	total := DumpLinksFromUrl(t, group, url, output_prefix+".0")

	jobs := make(chan [2]int, total/100+1)
	for i := 0; i < total/100; i++ {
		jobs <- [2]int{i*100 + 1, (i + 1) * 100}
	}
	if total > 100 && total%100 > 0 {
		jobs <- [2]int{total - total%100, total}
	}
	close(jobs)

	results := make(chan int)
	for i := 0; i < workers; i++ {
		go DownloadPageWorker(i, t, group, url, output_prefix, jobs, results)
	}
	for i := 0; i < workers; i++ {
		<-results
	}
}

func DownloadThreads(group string, workers int) {
	url := fmt.Sprintf("https://groups.google.com/forum/?_escaped_fragment_=forum/%s", group)
	output_prefix := fmt.Sprintf("%s/threads/t", group)
	fmt.Printf(":: Download topic from %s\n", group)
	DownloadPages("topic", group, url, output_prefix, workers)
}

func DownloadMessagesWorker(id int, group string, workers int, jobs <-chan string, results chan<- int) {
	output_prefix := fmt.Sprintf("%s/msgs/m", group)
	for url := range jobs {
		ss := strings.Split(url, "/")
		msg_id := ss[len(ss)-1]
		fmt.Printf(":: Downloading msg from %s/%s by worker[%d]\n", group, msg_id, id)
		DownloadPages("msg", group, url, fmt.Sprintf("%s.%s", output_prefix, msg_id), workers)
	}
	results <- 1
}

func DownloadMessages(group string, workers int) {
	dir := fmt.Sprintf("%s/threads/", group)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	jobs := make(chan string, 100)

	results := make(chan int)
	for i := 0; i < workers; i++ {
		go DownloadMessagesWorker(i, group, workers, jobs, results)
	}

	for _, f := range files {
		file, err := os.Open(fmt.Sprintf("%s/%s", dir, f.Name()))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()
			url := strings.Replace(text, "/d/topic/", "/forum/?_escaped_fragment_=topic/", -1)
			jobs <- url
		}
	}

	close(jobs)
	for i := 0; i < workers; i++ {
		<-results
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "./crawler <group-name> <worker-count>\n")
		return
	}
	group := os.Args[1]
	workers, _ := strconv.Atoi(os.Args[2])
	MkdirAll(group)
	DownloadThreads(group, workers)
	DownloadMessages(group, workers)
}
