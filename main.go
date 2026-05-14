package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func main() {

	// todo : move host to DOCKER_HOST env
	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}
	defer cli.Close()

	StartPostgresContainer(cli)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	slog.Info("server starting.....")

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func StartPostgresContainer(cli *client.Client) {

	slog.Info("starting database container....")

	res, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Hostname:     "database",
			Domainname:   "database",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			ExposedPorts: make(nat.PortSet),
			Cmd:          strslice.StrSlice{"ls"},
		},
		&container.HostConfig{AutoRemove: true},
		&network.NetworkingConfig{},
		&v1.Platform{Architecture: "x86_64"},
		"postgres",
	)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("starting database container....", "response", res.ID)
}
