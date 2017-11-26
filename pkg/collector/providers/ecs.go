// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017 Datadog, Inc.

// +build ecs

package providers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/collector/listeners"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/docker"
	log "github.com/cihub/seelog"
)

const (
	ecsADLabelPrefix        = "com.datadoghq.ad."
	metadataURL      string = "http://169.254.170.2/v2/metadata"
)

// ECSConfigProvider implements the ConfigProvider interface.
// It collects configuration templates from the ECS metadata API.
type ECSConfigProvider struct {
	client http.Client
}

// NewECSConfigProvider returns a new ECSConfigProvider.
// It configures an http Client with a 500 ms timeout.
func NewECSConfigProvider(config config.ConfigurationProviders) (ECSConfigProvider, error) {
	c := http.Client{
		Timeout: 500 * time.Millisecond,
	}
	return &ECSConfigProvider{
		client: c,
	}
}

// String returns a string representation of the ECSConfigProvider
func (p *ECSConfigProvider) String() string {
	return "ECS container labels"
}

// Collect finds all running containers in the agent's task, reads their labels
// and extract configuration templates from them for auto discovery.
func (p *ECSConfigProvider) Collect() ([]check.Config, error) {
	meta, err := p.getTaskMetadata()
	if err != nil {
		return nil, err
	}
	return p.parseECSContainers(meta.Containers)
}

// getTaskMetadata queries the ECS metadata API and unmarshals the resulting json
// into a TaskMetadata object.
func (p *ECSConfigProvider) getTaskMetadata() (listeners.TaskMetadata, error) {
	var meta listeners.TaskMetadata
	resp, err := p.client.Get(metadataURL)
	if err != nil {
		log.Errorf("unable to get task metadata - %s", err)
		return meta, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&meta)
	if err != nil {
		log.Errorf("unable to decode task metadata response - %s", err)
	}
	return meta, err
}

// parseECSContainers loops through ecs containers found in the ecs metadata response
// and extracts configuration templates out of their labels.
func (p *ECSConfigProvider) parseECSContainers(containers []listeners.ECSContainer) ([]check.Config, error) {
	var templates []check.Config
	for _, c := range containers {
		configs, err := extractTemplatesFromMap(docker.ContainerIDToEntityName(c.DockerID), c.Labels, ecsADLabelPrefix)
		switch {
		case err != nil:
			log.Errorf("unable to extract templates for container %s - %s", c.DockerID, err)
			continue
		case len(configs) == 0:
			continue
		default:
			templates = append(templates, configs...)
		}
	}
	return templates, nil
}