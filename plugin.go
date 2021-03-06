package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/pkg/spi/instance"
)

const (
	DefaultImage = "docker:1.12.3-dind"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Spec is a docker container run options.
type ContainerSpec struct {
	Name     string
	Hostname string
	Ports    []string
	Image    string
	Option   string
}

type plugin struct {
	client *DockerClient
	err    error
}

func NewDInDInstancePlugin() instance.Plugin {
	log.Debugln("dind instance plugin.")

	p := &plugin{}
	p.client, p.err = NewEnvDockerClient()

	return p
}

// Validate performs local validation on a provision request
func (p *plugin) Validate(req json.RawMessage) error {
	log.Debugln("validate", string(req))

	if p.err != nil {
		log.Error(p.err)
		return p.err
	}

	spec := ContainerSpec{}
	if err := json.Unmarshal(req, &spec); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// Provision creates a new instance based on the spec
func (p *plugin) Provision(spec instance.Spec) (*instance.ID, error) {
	log.Debugln("provision")
	if spec.Properties == nil {
		return nil, fmt.Errorf("no-properties")
	}

	cSpec := &ContainerSpec{
		Image: DefaultImage,
	}
	if err := json.Unmarshal(*spec.Properties, &cSpec); err != nil {
		log.Error(err)
		return nil, err
	}

	ctx := context.Background()

	image := cSpec.Image
	labels := spec.Tags

	idx := rand.Int31()
	var (
		name     string
		hostname string
	)

	if spec.LogicalID != nil {
		logicalID := string(*spec.LogicalID)
		name = logicalID
		hostname = logicalID
		labels["LogicalID"] = logicalID
		log.Debugln("logicalID", logicalID)
	} else {
		name = fmt.Sprintf("%s-%d", cSpec.Name, idx)
		hostname = fmt.Sprintf("%s-%d", cSpec.Hostname, idx)
	}

	log.Debugln("container", name, hostname, image, labels)

	containerID, err := p.client.ContainerRun(ctx, name, hostname, image, labels)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if spec.Init != "" {
		cmds := strings.Split(spec.Init, "\n")
		for _, cmd := range cmds {
			log.Debugln("cmd:", cmd)
			if err := p.client.ContainerExecCmd(ctx, containerID, cmd); err != nil {
				log.Error(err)
				return nil, err
			}
		}
	}

	id := instance.ID(containerID)
	return &id, nil

}

func (p *plugin) Destroy(instance instance.ID) error {
	log.Debugln("destroy")
	id := string(instance)
	ctx := context.Background()
	if err := p.client.ContainerStopAndRemove(ctx, id); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (p *plugin) DescribeInstances(tags map[string]string) ([]instance.Description, error) {
	log.Debugln("describeInstances")

	ctx := context.Background()

	containers, err := p.client.ContainerList(ctx, tags)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var descList []instance.Description

	for _, c := range containers {
		var logicalID instance.LogicalID
		if id, ok := c.Labels["LogicalID"]; ok {
			logicalID = instance.LogicalID(id)
		} else {
			logicalID = instance.LogicalID(c.ID)
		}
		log.Debugln("logicalID", logicalID)

		desc := instance.Description{
			ID:        instance.ID(c.ID),
			LogicalID: &logicalID,
			Tags:      c.Labels,
		}
		descList = append(descList, desc)
	}
	return descList, nil
}
