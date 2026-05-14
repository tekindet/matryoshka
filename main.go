package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Africa/Nairobi"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.AutoMigrate()

	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}
	defer cli.Close()

	// todo : to be attached to containers of the same mesh/swarm
	_, err = cli.NetworkCreate(
		context.Background(), "test-network",
		network.CreateOptions{
			Driver:     "bridge",
			Attachable: true,
		},
	)
	if err != nil {
		//log.Fatal(err)
		slog.Warn("could not create network", "error", err.Error())
	}

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
			"5435": struct{}{},
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
					HostPort: "5435",
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

	err = cli.ContainerStart(context.Background(), res.ID, container.StartOptions{})
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("container started successfully", "container", res.ID)

}
