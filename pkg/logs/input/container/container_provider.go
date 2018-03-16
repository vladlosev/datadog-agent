// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package container

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/docker/docker/api/types"
)

type Container struct {
	Identifier string
	LogSource  *config.LogSource
}

func NewContainer(container types.Container, source *config.LogSource) *Container {
	return &Container{}
}

type ContainerProvider struct {
	sources []*config.LogSource
}

func NewContainerProvider(sources []*config.LogSource) *ContainerProvider {
	return &ContainerProvider{
		sources: sources,
	}
}

func (p *ContainerProvider) ContainersToTail() ([]*Container, error) {
	containers, err := p.listContainers()
	if err != nil {
		return nil, err
	}
	var containersToTail []*Container
	for _, container := range containers {
		source := p.searchSourceForContainer(container)
		if source != nil {
			containersToTail = append(containersToTail, NewContainer(container.ID, source))
		}
	}
	return containers, nil
}

func (p *ContainerProvider) listContainers() ([]types.Container, error) {
	containers, err := s.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	return containers
}

func (p *ContainerProvider) searchSource(container types.Container) *config.LogSource {
	for _, source := range p.sources {

	}
	return false, nil
}

func (s *Scanner) shouldTailContainer(source *config.LogSource, container types.Container) bool {

}

func isImageMatch(imageFilter string, image string) bool {
	// Trim digest suffix if present
	splitted := strings.SplitN(image, digestPrefix, 2)
	image = splitted[0]
	// Expect repository to end with '/' or to be emty
	repository := strings.TrimSuffix(image, imageFilter)
	return len(repository) == 0 || strings.HasSuffix(repository, "/")
}

func isLabelMatch(labelFilter string, labels map[string]string) bool {
	// Expect a comma-separated list of label filters, eg: foo:bar, baz
	for _, value := range strings.Split(labelFilter, ",") {
		// Trim whitespace, then check whether the label format is either key:value or key=value
		label := strings.TrimSpace(value)
		parts := strings.FieldsFunc(label, func(c rune) bool {
			return c == ':' || c == '='
		})
		// If we have exactly two parts, check there is a container label that matches both.
		// Otherwise fall back to checking the whole label exists as a key.
		if _, exists := labels[label]; exists || len(parts) == 2 && labels[parts[0]] == parts[1] {
			return true
		}
	}
	return false
}
