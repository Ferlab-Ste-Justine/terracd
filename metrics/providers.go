package metrics

import (
	"os"
	"path/filepath"
)

type Provider struct {
	Registry string
	Organization string
	Name string
	Version string
}

func GetProvidersInfo(workDir string) ([]Provider, error) {
	result := []Provider{}

	providerDir := filepath.Join(workDir, ".terraform", "providers")

	registries, registriesErr := os.ReadDir(providerDir)
	if registriesErr != nil {
		return result, registriesErr
	}

	for _, registry := range registries {
		registryDir := filepath.Join(providerDir, registry.Name())
		
		organisations, organisationsErr := os.ReadDir(registryDir)
		if organisationsErr != nil {
			return result, organisationsErr
		}

		for _, organisation := range organisations {
			organisationDir := filepath.Join(registryDir, organisation.Name())

			providers, providersErr := os.ReadDir(organisationDir)
			if providersErr != nil {
				return result, providersErr
			}

			for _, provider := range providers {
				providerDir := filepath.Join(organisationDir, provider.Name())

				providerVersions, providerVersionsErr := os.ReadDir(providerDir)
				if providerVersionsErr != nil {
					return result, providerVersionsErr
				}

				for _, providerVersion := range providerVersions {
					result = append(result, Provider{
						Registry: registry.Name(),
						Organization: organisation.Name(),
						Name: provider.Name(),
						Version: providerVersion.Name(),
					})
				}
			}
		}
	}

	return result, nil
}