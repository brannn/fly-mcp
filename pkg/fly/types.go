package fly

import (
	"time"

	"github.com/superfly/fly-go"
)

// App represents a Fly.io application
type App struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Status       string                 `json:"status"`
	Deployed     bool                   `json:"deployed"`
	Hostname     string                 `json:"hostname"`
	AppURL       string                 `json:"appUrl"`
	Organization *fly.OrganizationBasic `json:"organization,omitempty"`
	CreatedAt    *time.Time             `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time             `json:"updatedAt,omitempty"`
}

// AppStatus represents the current status of an application
type AppStatus struct {
	AppName       string         `json:"appName"`
	Status        string         `json:"status"`
	Deployed      bool           `json:"deployed"`
	MachineCount  int            `json:"machineCount"`
	MachineStates map[string]int `json:"machineStates"`
	Hostname      string         `json:"hostname"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// Machine represents a Fly.io machine
type Machine struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	State    string            `json:"state"`
	Region   string            `json:"region"`
	ImageRef string            `json:"imageRef"`
	Config   *MachineConfig    `json:"config,omitempty"`
	Events   []MachineEvent    `json:"events,omitempty"`
	Checks   []MachineCheck    `json:"checks,omitempty"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// MachineConfig represents machine configuration
type MachineConfig struct {
	Image    string                 `json:"image"`
	Env      map[string]string      `json:"env,omitempty"`
	Services []MachineService      `json:"services,omitempty"`
	Mounts   []MachineMount        `json:"mounts,omitempty"`
	Guest    *MachineGuest         `json:"guest,omitempty"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// MachineService represents a service configuration
type MachineService struct {
	Protocol     string `json:"protocol"`
	InternalPort int    `json:"internalPort"`
	Ports        []Port `json:"ports,omitempty"`
}

// Port represents a port configuration
type Port struct {
	Port     int      `json:"port"`
	Handlers []string `json:"handlers,omitempty"`
}

// MachineMount represents a volume mount
type MachineMount struct {
	Volume      string `json:"volume"`
	Path        string `json:"path"`
	SizeGB      int    `json:"sizeGb,omitempty"`
	Encrypted   bool   `json:"encrypted,omitempty"`
}

// MachineGuest represents machine resource configuration
type MachineGuest struct {
	CPUKind  string `json:"cpuKind"`
	CPUs     int    `json:"cpus"`
	MemoryMB int    `json:"memoryMb"`
}

// MachineEvent represents a machine event
type MachineEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Request   map[string]interface{} `json:"request,omitempty"`
}

// MachineCheck represents a health check
type MachineCheck struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Output    string    `json:"output,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Volume represents a Fly.io volume
type Volume struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	State             string    `json:"state"`
	SizeGB            int       `json:"sizeGb"`
	Region            string    `json:"region"`
	Encrypted         bool      `json:"encrypted"`
	AttachedMachineID string    `json:"attachedMachineId,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Secret represents an application secret
type Secret struct {
	Name      string    `json:"name"`
	Digest    string    `json:"digest"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Region represents a Fly.io region
type Region struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// LogEntry represents a log entry from Fly.io
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Instance  string                 `json:"instance,omitempty"`
	Region    string                 `json:"region,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	Description string    `json:"description"`
	User        string    `json:"user,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ScaleConfig represents scaling configuration
type ScaleConfig struct {
	Count    int    `json:"count"`
	CPUKind  string `json:"cpuKind,omitempty"`
	CPUs     int    `json:"cpus,omitempty"`
	MemoryMB int    `json:"memoryMb,omitempty"`
}

// Certificate represents an SSL certificate
type Certificate struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	Type         string    `json:"type"`
	Source       string    `json:"source"`
	IsApex       bool      `json:"isApex"`
	IsWildcard   bool      `json:"isWildcard"`
	IsConfigured bool      `json:"isConfigured"`
	Check        bool      `json:"check"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Organization represents a Fly.io organization
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

// User represents a Fly.io user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AppConfig represents application configuration
type AppConfig struct {
	AppName     string                 `json:"appName"`
	PrimaryRegion string               `json:"primaryRegion,omitempty"`
	Env         map[string]string      `json:"env,omitempty"`
	Services    []ServiceConfig        `json:"services,omitempty"`
	Mounts      []MountConfig          `json:"mounts,omitempty"`
	Build       *BuildConfig           `json:"build,omitempty"`
	Deploy      *DeployConfig          `json:"deploy,omitempty"`
	Metrics     *MetricsConfig         `json:"metrics,omitempty"`
}

// ServiceConfig represents service configuration
type ServiceConfig struct {
	InternalPort int           `json:"internalPort"`
	Protocol     string        `json:"protocol"`
	Ports        []PortConfig  `json:"ports,omitempty"`
	Checks       []CheckConfig `json:"checks,omitempty"`
}

// PortConfig represents port configuration
type PortConfig struct {
	Port     int      `json:"port"`
	Handlers []string `json:"handlers,omitempty"`
}

// CheckConfig represents health check configuration
type CheckConfig struct {
	Type     string `json:"type"`
	Port     int    `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	Interval string `json:"interval,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Method   string `json:"method,omitempty"`
}

// MountConfig represents mount configuration
type MountConfig struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Type        string `json:"type,omitempty"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Dockerfile string            `json:"dockerfile,omitempty"`
	Image      string            `json:"image,omitempty"`
	Args       map[string]string `json:"args,omitempty"`
}

// DeployConfig represents deployment configuration
type DeployConfig struct {
	Strategy string `json:"strategy,omitempty"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Port int    `json:"port,omitempty"`
	Path string `json:"path,omitempty"`
}
