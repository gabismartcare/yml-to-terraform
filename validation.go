package main

import (
	"fmt"
	"strings"
)

func (s *ProjectConfiguration) IsValid() []error {
	errors := make([]error, 0, 10)
	serviceAccounts := s.getDeclaredServiceAccounts()
	errors = validateCloudRunDeclaration(s.GabiProject.CloudRuns, serviceAccounts, errors)
	errors = validatePubsubDeclaration(s.GabiProject.PubSub, serviceAccounts, errors)
	errors = validateRoles(s.GabiProject, errors)
	for k, v := range serviceAccounts {
		if !v {
			errors = append(errors, fmt.Errorf("- service account %s declared but not used", k))
		}
	}
	return errors
}

func validateRoles(project GabiProject, errors []error) []error {
	customRoleMaps := make(map[string]bool, len(project.CustomRole))
	for i := range project.CustomRole {
		customRoleMaps[project.CustomRole[i].Name] = true
	}
	for i := range project.ServiceAccounts {
		sa := project.ServiceAccounts[i]
		for j := range sa.Roles {
			role := sa.Roles[j]
			if !strings.HasPrefix(role, "roles") && !contains(customRoleMaps, role) {
				errors = append(errors, fmt.Errorf("- role %s has an invalid format", role))
			}
		}
	}
	return errors
}

func validatePubsubDeclaration(s PubSub, serviceAccounts map[string]bool, errors []error) []error {
	for i := range s.Topics {
		p := s.Topics[i]
		for j := range p.Subscriptions {
			ss := p.Subscriptions[j]
			if ss.ServiceAccount == "" || !contains(serviceAccounts, ss.ServiceAccount) {
				errors = append(errors, fmt.Errorf("- cloud run %s: service account not declared in gabi_project.service_accounts", ss.Name))
			} else {
				serviceAccounts[ss.ServiceAccount] = true
			}
		}
	}
	return errors
}

func validateCloudRunDeclaration(s []CloudRun, serviceAccounts map[string]bool, errors []error) []error {
	for i := range s {
		cr := s[i]
		if cr.ServiceAccount == "" || !contains(serviceAccounts, cr.ServiceAccount) {
			errors = append(errors, fmt.Errorf("- cloud run %s: service account not declared in gabi_project.service_accounts", cr.Name))
		} else {
			serviceAccounts[cr.ServiceAccount] = true
		}
		if cr.Name == "" {
			errors = append(errors, fmt.Errorf("- cloud run name not found"))
		}
		if cr.Location == "" {
			errors = append(errors, fmt.Errorf("- cloud run %s has no location not found", cr.Name))
		}
	}
	return errors
}

func (s *ProjectConfiguration) getDeclaredServiceAccounts() map[string]bool {
	serviceAccounts := make(map[string]bool, len(s.GabiProject.ServiceAccounts))
	for i := range s.GabiProject.ServiceAccounts {
		sc := s.GabiProject.ServiceAccounts[i]
		serviceAccounts[sc.Name] = sc.Orphan
	}
	return serviceAccounts
}
