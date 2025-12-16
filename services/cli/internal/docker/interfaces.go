package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// ClientInterface defines the contract for Docker operations
type ClientInterface interface {
	Close() error
	Ping(ctx context.Context) error
	GetVersion(ctx context.Context) (string, error)
	PullImage(ctx context.Context, imageName string) error
	CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, containerName string) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout *int) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	ListContainers(ctx context.Context, all bool, labelFilter map[string]string) ([]ContainerInfo, error)
	GetContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error)
	StreamContainerLogs(ctx context.Context, containerID string) error
	CreateNetwork(ctx context.Context, name string) (string, error)
	NetworkExists(ctx context.Context, name string) (bool, error)
	RemoveNetwork(ctx context.Context, name string) error
	RemoveContainerByName(ctx context.Context, name string) error
	ContainerAttach(ctx context.Context, containerID string, opts container.AttachOptions) (types.HijackedResponse, error)
}

// ManagerInterface defines the contract for container management
type ManagerInterface interface {
	SetupNetwork(ctx context.Context) error
	StartMCPServer(ctx context.Context, spec MCPServerSpec, workspacePath string) (string, error)
	StartAgent(ctx context.Context, spec AgentSpec, workspacePath string) (string, error)
	StopContainers(ctx context.Context) error
	CleanupContainers(ctx context.Context) error
	GetContainerStatus(ctx context.Context) ([]ContainerInfo, error)
}

// RunnerInterface defines the contract for native process execution
type RunnerInterface interface {
	Start(ctx context.Context, config RunnerConfig) error
	Run(ctx context.Context, config RunnerConfig, stdin io.Reader, stdout, stderr io.Writer) error
	Stop() error
	Wait() error
	IsRunning() bool
	PID() int
}

// Compile-time interface checks
var _ ClientInterface = (*Client)(nil)
var _ ManagerInterface = (*Manager)(nil)
var _ RunnerInterface = (*Runner)(nil)
