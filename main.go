package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/moss/mossspore/internal/spore"
	"github.com/moss/mossspore/internal/version"
)

func main() {
	var (
		configPath = flag.String("config", "", "path to config file")
		meshID     = flag.String("mesh-id", "", "mesh network identifier (overrides config)")
		listenPort = flag.Int("listen-port", 0, "listen port (0 = OS-assigned, overrides config)")
		verbose    = flag.Bool("v", false, "verbose logging")
		showVer    = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	cfg := resolveConfig(*configPath, *meshID, *listenPort, *verbose)

	if cfg.Verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(log.Ldate | log.Ltime)
	}

	log.Printf("starting %s", version.Info())

	sp, err := spore.New(cfg)
	if err != nil {
		log.Fatalf("fatal: %v", err)
	}

	if err := sp.Start(); err != nil {
		log.Fatalf("fatal: %v", err)
	}

	sp.WaitSignal()
}

// resolveConfig merges CLI flags with file-based config. CLI flags take
// precedence over file values.
func resolveConfig(configPath, meshID string, listenPort int, verbose bool) spore.Config {
	cfg := spore.DefaultConfig()

	if configPath != "" {
		raw, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("cannot read config %s: %v", configPath, err)
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			log.Fatalf("cannot parse config %s: %v", configPath, err)
		}
		log.Printf("loaded config from %s", configPath)
	}

	cfg.Normalize()

	if meshID != "" {
		cfg.MeshID = meshID
	}
	if listenPort > 0 {
		cfg.ListenPort = listenPort
	}
	cfg.Verbose = cfg.Verbose || verbose

	if err := validateConfig(cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	return cfg
}

func validateConfig(cfg spore.Config) error {
	// An empty mesh_id is valid: a substrate-only relay that serves every room.
	if strings.ContainsAny(cfg.MeshID, " \t\n") {
		return fmt.Errorf("mesh_id must not contain whitespace")
	}
	if cfg.IdentityPath != "" {
		dir := filepath.Dir(cfg.IdentityPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("cannot create identity directory %s: %w", dir, err)
		}
	}
	return nil
}
