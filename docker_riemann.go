package main

import (
	"flag"
	"log"
	"net/url"
	"os"

	"github.com/amir/raidman"
	"github.com/samalba/dockerclient"
)

type Config struct {
	Debug   bool
	Id      string
	Docker  string
	Riemann string
}

type Client struct {
	Docker  *dockerclient.DockerClient
	Riemann *raidman.Client
}

func connectToDocker(config *Config) (docker *dockerclient.DockerClient) {
	docker, err := dockerclient.NewDockerClient(config.Docker)
	if err != nil {
		log.Fatalf("Failed to connect to Docker (%s): %s", config.Docker, err)
	}

	if version, err := docker.Version(); err != nil {
		log.Fatalf("Failed to fetch docker daemon version: %s", err)
	} else {
		log.Printf("Connected to docker @ %s (version %s)", config.Docker, version.Version)
	}

	return
}

func connectToRiemann(config *Config) (riemann *raidman.Client) {
	url, err := url.Parse(config.Riemann)
	if err != nil {
		log.Fatalf("Failed to parse Riemann URL %s", config.Riemann)
	}

	if riemann, err = raidman.Dial(url.Scheme, url.Host); err != nil {
		log.Fatalf("Failed to connect to Riemann (%s): %s", config.Riemann, err)
	}

	return
}

func dockerEventCallback(event *dockerclient.Event, args ...interface{}) {
	config := args[0].(*Config)
	client := args[1].(*Client)

	riemann_item := raidman.Event{
		Description: event.Id,     // container id
		Host:        config.Id,    // command line provided id, or hostname
		Service:     event.Status, // create, start, die, destroy, ...
		Metric:      1,
		State:       "ok",
		Tags:        []string{"docker"},
	}

	if config.Debug {
		log.Printf("Sending event: %#v\n", riemann_item)
	}

	if err := client.Riemann.Send(&riemann_item); err != nil {
		log.Printf("! Error sending metric to Riemann: %s", err)
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}

func parseCommandLine() *Config {
	var config Config
	flag.BoolVar(&config.Debug, "debug", false,
		"Debug mode: outputs Riemann messages upon sending")
	flag.StringVar(&config.Id, "id", getHostname(),
		"Unique identifier used as Riemann events originating host")
	flag.StringVar(&config.Docker, "docker", "unix:///var/run/docker.sock",
		"Docker daemon location")
	flag.StringVar(&config.Riemann, "riemann", "tcp://localhost:5555",
		"Riemann service location")
	flag.Parse()
	return &config
}

func main() {
	config := parseCommandLine()

	// Connect to both Docker and Riemann.
	client := &Client{}
	client.Docker = connectToDocker(config)
	client.Riemann = connectToRiemann(config)
	defer func() { client.Riemann.Close() }()

	// Start to monitor events and enter an infinite select.
	client.Docker.StartMonitorEvents(dockerEventCallback, config, client)
	select {}
}
