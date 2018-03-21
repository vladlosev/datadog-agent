# Auto Discovery for logs

- Authors: Alexandre Jacquemot
- Date: 2018-03-21
- Status: draft

## Overview

We would like logs-agent to support the exact same configuration providers as the infra-agent to operate.

## Problem

Today, we have implemented custom solutions following the infra-agent guidelines and conventions to collect configurations from files and Docker labels. As a result, we duplicated some logic and we are maintaining two different solutions performing the same actions in the same code base, this is error prone and it degrades the user experience.

## Constraints

    1. All packages should converge to only one solution when it comes to retrieve configurations.
    2. Configurations should not be considered only as infra configurations.
    3. All configuration providers should be able to provide logs configuration.
    4. Agent status should be able to report config errors from different sources.
    5. A configuration must point to its resource (container identifier, pod name) and its type (file, docker labels, kubernetes annotations, ...).

## Recommended Solution

### Plug to Auto Discovery 

Auto Discovery is the unique solution to collect and provide configurations to components. Those components conform to an interface and implement their own logic to handle added and removed configurations. It is the responsibility of components to filter the configurations, Auto-Discovery is agnostic to providers and consumers. The same configuration is never provided twice except if it has been removed and added back.

![high level design](https://github.com/DataDog/datadog-agent/tree/ajacquemot/logs_ad_proposal/docs/proposals/metadata/ad_logs.png "High Level Design")

## Other Solutions

###  Custom Configurations

Keep developing and maintaining custom configuration solutions for logs-specific needs sharing and reusing as much code as possible with the infra-agent.

## Open Questions

- Is there any other team that would be interested in reusing Auto Discovery ?
- Should we split the auto-discovery logic from the check logic ?

## Appendix

- https://github.com/DataDog/datadog-agent/tree/master/pkg/logs
- https://github.com/DataDog/datadog-agent/tree/master/pkg/collector/autodiscovery
- https://docs.google.com/drawings/d/1jwjygOhmII4EPiTw5v69lnC1UC7Ah0g9FDPl0irdSyE
