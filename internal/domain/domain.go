package domain

const (
	ServiceTypeApp      = "app"
	ServiceTypePostgres = "postgres"
	ServiceTypeRedis    = "redis"
	ServiceTypeQueue    = "queue"
	ServiceTypeWorker   = "worker"
)

const (
	ServiceStatusPending = "pending"
	ServiceStatusRunning = "running"
	ServiceStatusStopped = "stopped"
	ServiceStatusFailed  = "failed"
)

type Project struct {
	ID          string `gorm:"id" json:"id"`
	Name        string `gorm:"name" json:"name"`
	Description string `gorm:"description" json:"description"`
}

type Service struct {
	ID        string `gorm:"id" json:"id"`
	ProjectID string `gorm:"project_id" json:"project_id"`
	Name      string `gorm:"name" json:"name"`
	Type      string `gorm:"column:type" json:"type"`
	Status    string `gorm:"status" json:"status"`
}

type Network struct {
	ID        string `gorm:"id" json:"id"`
	ProjectID string `gorm:"project_id" json:"project_id"`
	Name      string `gorm:"name" json:"name"`
}

type Deployment struct {
	ID        string `gorm:"id" json:"id"`
	ServiceID string `gorm:"service_id" json:"service_id"`
}
