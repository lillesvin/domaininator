package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp/syntax"
	"strings"
	"sync"
	"syscall"

	"github.com/alixaxel/genex"
	"github.com/cheggaaa/pb"
)

var (
	AppName    string = "Domaininator"
	AppVersion string = "0.0.1"
	workers    int
	version    bool
	dbg        bool
)

func init() {
	flag.IntVar(&workers, "workers", 4, "Number of parallel workers to run")
	flag.BoolVar(&version, "version", false, "Show version info and exit")
	flag.Parse()
}

func main() {
	if version {
		fmt.Printf("%s  %s\n", AppName, AppVersion)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	var waitGroup sync.WaitGroup
	var outBuffer []string
	var count int

	workerKill := make(chan bool)
	queueKill := make(chan bool)
	lookupChan := make(chan string)
	responseChan := make(chan string)

	interruptHandler(workerKill, queueKill, workers)

	args := flag.Args()

	charset, _ := syntax.Parse(`[a-z0-9]`, syntax.Perl)
	input, err := syntax.Parse(args[0], syntax.Perl)
	if err != nil {
		fmt.Printf("Error while parsing pattern: %s\n", err)
		os.Exit(2)
	}
	count = int(genex.Count(input, charset, 3))

	bar := pb.StartNew(count)

	go func() {
		for res := range responseChan {
			outBuffer = append(outBuffer, res)
		}
	}()

	for i := 0; i < workers; i++ {
		go worker(&waitGroup, lookupChan, responseChan, workerKill, bar)
		waitGroup.Add(1)
	}

	var domains []string
	genex.Generate(input, charset, 3, func(domain string) {
		domains = append(domains, domain)
	})

	for _, domain := range domains {
		select {
		case <-queueKill:
			fmt.Println("Closing lookup queue...")
			close(lookupChan)
			return
		default:
			lookupChan <- domain
		}
	}
	close(lookupChan)

	bar.Finish()
	waitGroup.Wait()

	fmt.Println("\n")

	// Print output
	fmt.Println(strings.Join(outBuffer, "\n"))
}

func worker(wg *sync.WaitGroup, lookups chan string, responses chan string, workerKill chan bool, bar *pb.ProgressBar) {
	defer wg.Done()

	for d := range lookups {
		select {
		case <-workerKill:
			fmt.Println("Stopping worker")
			return
		default:
			dns := DNSLookup(d)
			if len(dns) > 0 {
				responses <- fmt.Sprintf("%s: %s", d, strings.Join(dns, ","))
			}
			bar.Increment()
		}
	}
}

func interruptHandler(workerKill, queueKill chan bool, workers int) {
	interruptChan := make(chan os.Signal)

	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interruptChan

		fmt.Println("\nInterrupt caught. Cleaning up...")
		queueKill <- true
		for i := 0; i < workers; i++ {
			workerKill <- true
		}
	}()
}

func DNSLookup(domain string) []string {
	var ret []string

	ips, _ := net.LookupHost(domain)
	if len(ips) > 0 {
		ret = append(ret, "A")
	}

	mxs, _ := net.LookupMX(domain)
	if len(mxs) > 0 {
		ret = append(ret, "MX")
	}

	nss, _ := net.LookupNS(domain)
	if len(nss) > 0 {
		ret = append(ret, "NS")
	}

	return ret
}