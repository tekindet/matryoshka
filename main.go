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
	"github.com/tekindet/matryoshka/internal/domain"
	"github.com/tekindet/matryoshka/internal/graphql"
	"github.com/tekindet/matryoshka/internal/manager"
	"github.com/tekindet/matryoshka/internal/orchestrator"
	"github.com/tekindet/matryoshka/internal/store"

	"github.com/joho/godotenv"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Africa/Nairobi"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.AutoMigrate(&domain.Project{}, &domain.Service{})

	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatalf("docker client init failed : %v", err)
	}
	defer cli.Close()

	ptStore := store.NewPostgresStore(db)
	dockerOrch := orchestrator.NewDockerOrchestrator(cli)
	paasManager := manager.NewPaaSManager(ptStore, dockerOrch)

	demoProj, err := paasManager.CreateProject(context.Background(), "billing-system", "Handles client subscriptions")
	if err != nil {
		slog.Error("Failed to seed demo project", "error", err)
	} else {
		slog.Info("Successfully bootstrapped isolated project", "project_id", demoProj.ID, "name", demoProj.Name)
	}

	//StartPostgresContainer(cli)

	gqlResolver := graphql.New(paasManager)
	http.Handle("/graphql", gqlResolver.Handler())

	http.HandleFunc("/health", HealthCheckHandler)
	slog.Info("Matryoshka PaaS Engine started running on port :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
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
