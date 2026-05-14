package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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

	cnt_cfg := &container.Config{
		Image:        "postgres:latest",
		Hostname:     "database",
		Domainname:   "database",
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: nat.PortSet{
			"5432": struct{}{},
		},
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
		},
	}

	hst_cfg := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: container.NetworkMode("bridge"),
		PortBindings: nat.PortMap{
			"5432/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "5432",
				},
			},
		},
	}

	res, err := cli.ContainerCreate(
		context.Background(),
		cnt_cfg,
		hst_cfg,
		&network.NetworkingConfig{},
		&v1.Platform{Architecture: "x86_64"},
		"postgres",
	)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("starting database container....", "response", res.ID)
}
