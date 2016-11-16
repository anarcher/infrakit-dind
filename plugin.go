package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker_client "github.com/docker/docker/client"
	"github.com/docker/infrakit/pkg/spi/instance"
)

const (
	DefaultImage      = "docker:1.12.3-dind"
	DefaultSpecOption = "-d --privileged "
)

// Spec is a docker container run options.
type ContainerSpec struct {
	Name     string
	Hostname string
	Ports    []string
	Image    string
	Option   string
}

type plugin struct {
	cli          *docker_client.Client
	err          error
	containerIDs []string
}

func NewDInDInstancePlugin() instance.Plugin {
	log.Debugln("dind instance plugin.")

	p := &plugin{}
	p.cli, p.err = docker_client.NewEnvClient()

	return p
}

// Validate performs local validation on a provision request
func (p *plugin) Validate(req json.RawMessage) error {
	log.Debugln("validate", string(req))

	if p.err != nil {
		return p.err
	}

	spec := ContainerSpec{}
	if err := json.Unmarshal(req, &spec); err != nil {
		return err
	}

	return nil
}

// Provision creates a new instance based on the spec
func (p *plugin) Provision(spec instance.Spec) (*instance.ID, error) {
	if spec.Properties == nil {
		return nil, fmt.Errorf("no-properties")
	}

	containerSpec := &ContainerSpec{
		Image: DefaultImage,
	}
	if err := json.Unmarshal(*spec.Properties, &containerSpec); err != nil {
		return nil, err
	}

	ctx := context.Background()

	cfg := &container.Config{
		Hostname: containerSpec.Hostname,
		Image:    containerSpec.Image,
	}
	hostCfg := &container.HostConfig{
		Privileged: true,
	}
	netCfg := &network.NetworkingConfig{}
	resp, err := p.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, containerSpec.Name)
	if err != nil {
		return nil, err
	}
	if err := p.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	p.containerIDs = append(p.containerIDs, resp.ID)

	//spec.Init string
	//docker exec ...

	id := instance.ID(resp.ID)
	return &id, nil

}

func (p *plugin) Destroy(instance instance.ID) error {
	id := string(instance)
	ctx := context.Background()
	d := 5 * time.Second
	if err := p.cli.ContainerStop(ctx, id, &d); err != nil {
		return err
	}
	if err := p.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func (p *plugin) DescribeInstances(tags map[string]string) ([]instance.Description, error) {
	//Get p.containerIDs and then show docker container info
	return []instance.Description{}, nil
}
