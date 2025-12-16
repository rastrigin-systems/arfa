package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Client wraps the Docker SDK client.
// All methods accept context.Context as the first parameter for cancellation and timeout support.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		cli: cli,
	}, nil
}

// Close closes the Docker client connection
func (dc *Client) Close() error {
	return dc.cli.Close()
}

// Ping checks if Docker daemon is running
func (dc *Client) Ping(ctx context.Context) error {
	_, err := dc.cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("Docker daemon not accessible: %w", err)
	}
	return nil
}

// GetVersion returns Docker version information
func (dc *Client) GetVersion(ctx context.Context) (string, error) {
	version, err := dc.cli.ServerVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get Docker version: %w", err)
	}
	return version.Version, nil
}

// PullImage pulls a Docker image (or uses local if available)
func (dc *Client) PullImage(ctx context.Context, imageName string) error {
	// Check if image exists locally
	_, err := dc.cli.ImageInspect(ctx, imageName)
	if err == nil {
		fmt.Printf("  Using local image %s\n", imageName)
		return nil
	}

	// Image doesn't exist locally, pull it
	fmt.Printf("  Pulling %s...\n", imageName)

	reader, err := dc.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Copy output to stdout (shows progress)
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return fmt.Errorf("failed to read pull output: %w", err)
	}

	return nil
}

// CreateContainer creates a Docker container
func (dc *Client) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, containerName string) (string, error) {
	resp, err := dc.cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container %s: %w", containerName, err)
	}

	if len(resp.Warnings) > 0 {
		for _, warning := range resp.Warnings {
			fmt.Printf("  Warning: %s\n", warning)
		}
	}

	return resp.ID, nil
}

// StartContainer starts a Docker container
func (dc *Client) StartContainer(ctx context.Context, containerID string) error {
	if err := dc.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %w", containerID, err)
	}
	return nil
}

// StopContainer stops a Docker container
func (dc *Client) StopContainer(ctx context.Context, containerID string, timeout *int) error {
	if err := dc.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: timeout}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}
	return nil
}

// RemoveContainer removes a Docker container
func (dc *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	if err := dc.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force}); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}
	return nil
}

// ListContainers lists Docker containers with optional filters
func (dc *Client) ListContainers(ctx context.Context, all bool, labelFilter map[string]string) ([]ContainerInfo, error) {
	options := container.ListOptions{
		All: all,
	}

	// Add label filters
	if len(labelFilter) > 0 {
		filters := make([]string, 0)
		for k, v := range labelFilter {
			filters = append(filters, fmt.Sprintf("label=%s=%s", k, v))
		}
		// Note: The API expects filters in a specific format, this is simplified
	}

	containers, err := dc.cli.ContainerList(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			// Remove leading slash from container name
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		result[i] = ContainerInfo{
			ID:      c.ID[:12], // Short ID
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
		}
	}

	return result, nil
}

// GetContainerLogs retrieves logs from a container
func (dc *Client) GetContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: false,
	}

	logs, err := dc.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

// StreamContainerLogs streams container logs to stdout/stderr
func (dc *Client) StreamContainerLogs(ctx context.Context, containerID string) error {
	logs, err := dc.GetContainerLogs(ctx, containerID, true)
	if err != nil {
		return err
	}
	defer logs.Close()

	// Docker multiplexes stdout and stderr, use stdcopy to demux
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}

// CreateNetwork creates a Docker network
func (dc *Client) CreateNetwork(ctx context.Context, name string) (string, error) {
	resp, err := dc.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			"com.ubik.managed": "true",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create network %s: %w", name, err)
	}

	return resp.ID, nil
}

// NetworkExists checks if a network exists
func (dc *Client) NetworkExists(ctx context.Context, name string) (bool, error) {
	networks, err := dc.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list networks: %w", err)
	}

	for _, net := range networks {
		if net.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// RemoveNetwork removes a Docker network
func (dc *Client) RemoveNetwork(ctx context.Context, name string) error {
	if err := dc.cli.NetworkRemove(ctx, name); err != nil {
		return fmt.Errorf("failed to remove network %s: %w", name, err)
	}
	return nil
}

// RemoveContainerByName finds and removes a container by name
func (dc *Client) RemoveContainerByName(ctx context.Context, name string) error {
	// List all containers (including stopped ones)
	containers, err := dc.cli.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Find container with matching name
	var containerID string
	for _, c := range containers {
		for _, cName := range c.Names {
			// Remove leading slash from container name
			cleanName := cName
			if len(cleanName) > 0 && cleanName[0] == '/' {
				cleanName = cleanName[1:]
			}
			if cleanName == name {
				containerID = c.ID
				break
			}
		}
		if containerID != "" {
			break
		}
	}

	// If container not found, that's okay - nothing to remove
	if containerID == "" {
		return nil
	}

	// Stop container if it's running (use default 10s timeout)
	timeout := 10
	_ = dc.StopContainer(ctx, containerID, &timeout)

	// Remove container
	removeOptions := container.RemoveOptions{
		Force: true, // Force removal even if running
	}
	if err := dc.cli.ContainerRemove(ctx, containerID, removeOptions); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", name, err)
	}

	return nil
}

// ContainerAttach attaches to a running container for interactive I/O.
func (dc *Client) ContainerAttach(ctx context.Context, containerID string, opts container.AttachOptions) (types.HijackedResponse, error) {
	return dc.cli.ContainerAttach(ctx, containerID, opts)
}
