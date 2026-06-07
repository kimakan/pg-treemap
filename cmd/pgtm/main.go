package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/kimakan/pg-treemap/internal/collector"
	"github.com/kimakan/pg-treemap/internal/config"
	"github.com/kimakan/pg-treemap/web"
)

func main() {
	configPath := flag.String("conf", config.DefaultConfigFile, "path to the config file")
	createConfig := flag.Bool("create-conf", false, "create default config in the current directory")
	collect := flag.Bool("collect", false, "collect metadata from the databases defined in the config and store it in the snapshot.json")
	serve := flag.Bool("serve", false, "serve web UI using the address defined in the config file")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *createConfig {
		err := config.CreateConfigFile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot create the config file: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load the config file: %v\n", err)
		os.Exit(1)
	}

	if *collect {
		err := collector.Collect(conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to collect the data: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, "Successfully collected the metadata")
		os.Exit(0)
	}
	if *serve {
		http.Handle("/", http.FileServer(http.FS(web.FS)))
		http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
			data, err := os.ReadFile("snapshot.json")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		})
		http.ListenAndServe(conf.ServeAddr, nil)
		os.Exit(0)
	}
	// default:
	fmt.Fprintf(os.Stderr, "unknown flag: %s\n\n", flag.Arg(0))
	flag.Usage()
	os.Exit(1)
}
