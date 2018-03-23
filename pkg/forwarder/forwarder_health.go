package forwarder

import (
	"expvar"
	"fmt"
	"net/http"
	"time"

	log "github.com/cihub/seelog"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/status/health"
)

var (
	apiKeyStatusUnknown = expvar.String{}
	apiKeyInvalid       = expvar.String{}
	apiKeyValid         = expvar.String{}
)

func init() {
	apiKeyStatusUnknown.Set("Unable to validate API Key")
	apiKeyInvalid.Set("API Key invalid")
	apiKeyValid.Set("API Key valid")
}

// forwarderHealth report the health status of the Forwarder. A Forwarder is
// unhealthy if the API keys are not longer valid or if to many transactions
// were dropped
type forwarderHealth struct {
	health *health.Handle
	stop   chan struct{}
	ddURL  string
}

func (fh *forwarderHealth) init() {
	fh.stop = make(chan struct{})
	fh.ddURL = config.Datadog.GetString("dd_url")
}

func (fh *forwarderHealth) Start(keysPerDomains map[string][]string) {
	fh.health = health.Register("forwarder")
	fh.init()
	go fh.healthCheckLoop(keysPerDomains)
}

func (fh *forwarderHealth) Stop() {
	fh.health.Deregister()
	close(fh.stop)
}

func (fh *forwarderHealth) healthCheckLoop(keysPerDomains map[string][]string) {
	log.Debug("Checking APIkey validity.")

	valid := fh.hasValidAPIKey(keysPerDomains)
	// If no key is valid, no need to keep checking, they won't magicaly become valid
	if !valid {
		log.Errorf("No valid api key found, reporting the forwarder as unhealthy.")
		return
	}

	validateTicker := time.NewTicker(time.Hour * 1)
	defer validateTicker.Stop()

	for {
		select {
		case <-fh.stop:
			return
		case <-validateTicker.C:
			valid := fh.hasValidAPIKey(keysPerDomains)
			if !valid {
				log.Errorf("No valid api key found, reporting the forwarder as unhealthy.")
				return
			}
		case <-fh.health.C:
			if transactionsExpvar.Get("DroppedOnInput") != nil && transactionsExpvar.Get("DroppedOnInput").String() != "0" {
				log.Errorf("Detected dropped transaction, reporting the forwarder as unhealthy: %v.", transactionsExpvar.Get("DroppedOnInput"))
				return
			}
		}
	}
}

func (fh *forwarderHealth) setAPIKeyStatus(apiKey string, domain string, status expvar.Var) {
	obfuscatedKey := fmt.Sprintf("%s,*************************", domain)
	if len(apiKey) > 5 {
		obfuscatedKey += apiKey[len(apiKey)-5:]
	}
	apiKeyStatus.Set(obfuscatedKey, status)
}

func (fh *forwarderHealth) validateAPIKey(apiKey, domain string) (bool, error) {
	url := fmt.Sprintf("%s%s?api_key=%s", fh.ddURL, v1ValidateEndpoint, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fh.setAPIKeyStatus(apiKey, domain, &apiKeyStatusUnknown)
		return false, err
	}
	defer resp.Body.Close()

	// Server will respond 200 if the key is valid or 403 if invalid
	if resp.StatusCode == 200 {
		fh.setAPIKeyStatus(apiKey, domain, &apiKeyValid)
		return true, nil
	} else if resp.StatusCode == 403 {
		fh.setAPIKeyStatus(apiKey, domain, &apiKeyInvalid)
		return false, nil
	}

	fh.setAPIKeyStatus(apiKey, domain, &apiKeyStatusUnknown)
	return false, fmt.Errorf("Unexpected response code from the apikey validation endpoint: %v", resp.StatusCode)
}

func (fh *forwarderHealth) hasValidAPIKey(keysPerDomains map[string][]string) bool {
	validKey := false
	apiError := false

	for domain, apiKeys := range keysPerDomains {
		for _, apiKey := range apiKeys {
			v, err := fh.validateAPIKey(apiKey, domain)
			if err != nil {
				log.Debug(err)
				apiError = true
			} else if v {
				validKey = true
			}
		}
	}

	// If there is an error during the api call, we assume that there is a
	// valid key to avoid killing lots of agent on an outage.
	if apiError {
		return true
	}
	return validKey
}
