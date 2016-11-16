package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	docker "github.com/docker/docker/api/types"
	docker_client "github.com/docker/docker/client"
	"github.com/docker/infrakit/pkg/spi/instance"
)

const (
	DefaultImage      = "docker:1.12.3-dind"
	DefaultSpecOption = "-d --privileged "
)

// Spec is a docker container run options.
type Spec struct {
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

	if p.err {
		return p.err
	}

	spec := Spec{}
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

	spec := Spec{
		Image: DefaultImage,
	}
	if err := json.Unmarshal(*spec.Properties, &spec); err != nil {
		return nil, err
	}

	ctx := context.Background()

	cfg := &docker.Config{
		Hostname: spec.Hostname,
		Image:    spec.Image,
	}
	hostCfg := &docker.HostConfig{
		Privileged: true,
	}
	netCfg := &docker.NetworkingConfig{}
	resp, err := p.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, spec.Name)
	if err != nil {
		return nil, err
	}
	if err := p.cli.ContainerStart(ctx, resp.ID, &docker.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	p.containerIDs = append(p.containerIDs, resp.ID)

	//spec.Init string
	//docker exec ...

	return instance.ID(resp.ID), nil

}

func (p *plugin) Destroy(instance instance.ID) error {
	ctx = context.Background()
	if err := p.cli.ContainerStop(ctx, instance, 1*time.Seconds); err != nil {
		return err
	}
	if err := p.cli.ContainerRemove(ctx, instance, &docker.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func (p *plugin) DescribeInstances(tags map[string]string) ([]instance.Description, error) {
	//Get p.containerIDs and then show docker container info
	return []instance.Description{}, nil
}
