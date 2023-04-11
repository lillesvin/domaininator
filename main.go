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
	AppName     string = "Domaininator"
	AppVersion  string = "0.1.0"
	version     bool
	cfg         *Config
	cfgFile     string
	flagShowIPs bool
	flagVerbose bool
	flagWorkers int
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [pattern]\n\n", os.Args[0])

		flag.PrintDefaults()
	}

	flag.StringVar(&cfgFile, "config", "", "Config file to use")
	flag.BoolVar(&flagShowIPs, "ip", false, "Show IPs on resolving domains")
	flag.BoolVar(&flagVerbose, "verbose", false, "Show all domain names, even if they are not registered")
	flag.BoolVar(&version, "version", false, "Show version info and exit")
	flag.IntVar(&flagWorkers, "workers", 16, "Number of parallel workers to run")
	flag.Parse()
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	// Print version info and exit
	if version {
		fmt.Printf("%s %s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// Config supplied? Otherwise try to find a default
	if cfgFile == "" {
		cfgFile, _ = FindConfig()
	}

	// Load config from file
	cfg, err := NewFromTOML(cfgFile)
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		os.Exit(2)
	}

	// No pattern in args and config?
	if flag.NArg() != 1 && cfg.Pattern == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Let flags override config settings
	if isFlagSet("ip") {
		cfg.ShowIPs = flagShowIPs
	}
	if isFlagSet("verbose") {
		cfg.Verbose = flagVerbose
	}
	if isFlagSet("workers") {
		cfg.Workers = flagWorkers
	}

	var (
		waitGroup sync.WaitGroup
		outBuffer []string
		count     int
	)

	// All the channels
	workerKill := make(chan bool)
	queueKill := make(chan bool)
	lookupChan := make(chan string, 16)
	responseChan := make(chan string)

	// Set up interrups handling
	interruptHandler(workerKill, queueKill, cfg.Workers)

	// Find the pattern, we're working with
	// Prefer command line pattern
	args := flag.Args()
	var pattern string
	if len(args) > 0 {
		pattern = args[0]
	} else {
		pattern = cfg.Pattern
	}

	fmt.Printf("Pattern: %s\n", pattern)
	fmt.Printf("Config: %s\n", cfgFile)

	// Setup genex
	charset, _ := syntax.Parse(`[a-z0-9]`, syntax.Perl)
	input, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		fmt.Printf("Error while parsing pattern: %s\n", err)
		os.Exit(2)
	}
	count = int(genex.Count(input, charset, 3))

	// Progress bar
	bar := pb.New(count)
	bar.Output = os.Stderr
	bar.Start()

	// Handle output
	go func() {
		for res := range responseChan {
			outBuffer = append(outBuffer, res)
		}
	}()

	// Start workers
	for i := 0; i < cfg.Workers; i++ {
		go worker(&waitGroup, lookupChan, responseChan, workerKill, bar, cfg)
		waitGroup.Add(1)
	}

	// Generate domain names
	var domains []string
	genex.Generate(input, charset, 3, func(domain string) {
		domains = append(domains, domain)
	})

	for _, domain := range domains {
		select {
		case <-queueKill:
			close(lookupChan)
			return
		default:
			lookupChan <- domain
		}
	}
	close(lookupChan)

	waitGroup.Wait()
	bar.Finish()

	// Print output
	fmt.Println("")
	fmt.Println(strings.Join(outBuffer, "\n"))
}

func worker(wg *sync.WaitGroup, lookups chan string, responses chan string, workerKill chan bool, bar *pb.ProgressBar, cfg *Config) {
	defer wg.Done()

	for d := range lookups {
		select {
		case <-workerKill:
			fmt.Println("Stopping worker")
			return
		default:
			if cfg.InWhitelist(d) == false {
				dns := DNSLookup(d, cfg)
				if len(dns) > 0 {
					responses <- fmt.Sprintf("%s: %s", d, strings.Join(dns, "; "))
				} else if cfg.Verbose {
					responses <- d
				}
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

func DNSLookup(domain string, cfg *Config) []string {
	var ret []string

	ips, _ := net.LookupHost(domain)
	if len(ips) > 0 {
		if cfg.ShowIPs {
			ret = append(ret, fmt.Sprintf("A: %s", strings.Join(ips, ", ")))
		} else {
			ret = append(ret, "A")
		}
	}

	mxs, _ := net.LookupMX(domain)
	if len(mxs) > 0 {
		if cfg.ShowIPs {
			var out []string
			for _, mx := range mxs {
				out = append(out, mx.Host)
			}
			ret = append(ret, fmt.Sprintf("MX: %s", strings.Join(out, ", ")))
		} else {
			ret = append(ret, "MX")
		}
	}

	nss, _ := net.LookupNS(domain)
	if len(nss) > 0 {
		if cfg.ShowIPs {
			var out []string
			for _, ns := range nss {
				out = append(out, ns.Host)
			}
			ret = append(ret, fmt.Sprintf("NS: %s", strings.Join(out, ", ")))
		} else {
			ret = append(ret, "NS")
		}
	}

	return ret
}
