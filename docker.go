package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/network"
	/*
		//client is newer than server (client API version: 1.25, server API version: 1.24)
		"github.com/docker/docker/api/types"
		"github.com/docker/docker/api/types/container"
		"github.com/docker/docker/api/types/filters"
		"github.com/docker/docker/api/types/network"
		"github.com/docker/docker/client"
	*/)

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
		Cmd:        strings.Split(cmd, " "),
		Privileged: true,
	}
	exec, err := c.client.ContainerExecCreate(ctx, id, cfg)
	if err != nil {
		return fmt.Errorf("failed to create exec: %v", err)
	}

	execCfg := types.ExecStartCheck{}
	if err := c.client.ContainerExecStart(ctx, exec.ID, execCfg); err != nil {
		return fmt.Errorf("failed to start exec: %v", err)

	}

	inspect, err := c.client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return fmt.Errorf("failed to check prepare command exit code: %v", err)
	}

	if c := inspect.ExitCode; c != 0 {
		return fmt.Errorf("prepare command exited with code %v", c)
	}

	return nil
}

func (c *DockerClient) ContainerList(ctx context.Context, tags map[string]string) ([]types.Container, error) {
	args := filters.NewArgs()
	for k, v := range tags {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	opts := types.ContainerListOptions{
		Filter: args,
	}

	ls, err := c.client.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	return ls, nil
}
