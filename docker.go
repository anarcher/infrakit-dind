package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	client *client.Client
}

func NewEnvDockerClient() (*DockerClient, error) {

	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	c := &DockerClient{
		client: client,
	}

	return c, nil
}

func (c *DockerClient) ContainerRun(ctx context.Context, name, hostname, image string, labels map[string]string) (string, error) {

	cfg := &container.Config{
		Hostname: hostname,
		Image:    image,
		Labels:   labels,
	}
	hostCfg := &container.HostConfig{
		Privileged: true,
	}
	netCfg := &network.NetworkingConfig{}

	resp, err := c.client.ContainerCreate(ctx, cfg, hostCfg, netCfg, name)
	if err != nil {
		return "", err
	}
	if err := c.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return resp.ID, err
}

func (c *DockerClient) ContainerStopAndRemove(ctx context.Context, id string) error {
	d := 5 * time.Second
	if err := c.client.ContainerStop(ctx, id, &d); err != nil {
		return err
	}
	if err := c.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *DockerClient) ContainerExecCmd(ctx context.Context, id string, cmd string) error {
	cfg := types.ExecConfig{
		Cmd: []string{cmd},
	}
	resp, err := c.client.ContainerExecCreate(ctx, id, cfg)
	if err != nil {
		return err
	}

	execCfg := types.ExecStartCheck{}
	if err := c.client.ContainerExecStart(ctx, resp.ID, execCfg); err != nil {
		return err
	}

	return nil
}

func (c *DockerClient) ContainerList(ctx context.Context, tags map[string]string) ([]types.Container, error) {
	args := filters.NewArgs()
	for k, v := range tags {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	opts := types.ContainerListOptions{
		Filters: args,
	}

	ls, err := c.client.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	return ls, nil
}
