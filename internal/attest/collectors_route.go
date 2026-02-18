package attest

import (
	"fmt"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type RouteConfig struct {
	RouteConfig      string                `yaml:"route_config"`
	ProviderSet      []types.ProviderModel `yaml:"provider_set"`
	BudgetPolicy     string                `yaml:"budget_policy"`
	FallbackGraph    string                `yaml:"fallback_graph"`
	RoutingStrategy  string                `yaml:"routing_strategy"`
	CanaryConfig     string                `yaml:"canary_config"`
	SimulationResult string                `yaml:"simulation_result"`
}

func CollectRoute(configPath string) (types.Statement, error) {
	cfg := RouteConfig{}
	if err := LoadConfig(configPath, &cfg); err != nil {
		return types.Statement{}, err
	}
	cfg.RouteConfig = resolvePath(configPath, cfg.RouteConfig)
	cfg.BudgetPolicy = resolvePath(configPath, cfg.BudgetPolicy)
	cfg.FallbackGraph = resolvePath(configPath, cfg.FallbackGraph)
	cfg.CanaryConfig = resolvePath(configPath, cfg.CanaryConfig)
	cfg.SimulationResult = resolvePath(configPath, cfg.SimulationResult)
	for _, req := range []struct {
		path string
		name string
	}{{cfg.RouteConfig, "route_config"}, {cfg.BudgetPolicy, "budget_policy"}, {cfg.FallbackGraph, "fallback_graph"}} {
		if err := requirePath(req.path, req.name); err != nil {
			return types.Statement{}, err
		}
	}
	if cfg.RoutingStrategy == "" {
		return types.Statement{}, fmt.Errorf("routing_strategy is required")
	}
	if len(cfg.ProviderSet) == 0 {
		return types.Statement{}, fmt.Errorf("provider_set is required")
	}

	routeDigest, _, _ := hash.DigestFile(cfg.RouteConfig)
	budgetDigest, _, _ := hash.DigestFile(cfg.BudgetPolicy)
	fallbackDigest, _, _ := hash.DigestFile(cfg.FallbackGraph)

	predicate := types.RoutePredicate{
		RouteConfigDigest:   routeDigest,
		ProviderSet:         cfg.ProviderSet,
		BudgetPolicyDigest:  budgetDigest,
		FallbackGraphDigest: fallbackDigest,
		RoutingStrategy:     cfg.RoutingStrategy,
	}
	if cfg.CanaryConfig != "" {
		d, _, err := hash.DigestFile(cfg.CanaryConfig)
		if err != nil {
			return types.Statement{}, err
		}
		predicate.CanaryConfigDigest = d
	}
	if cfg.SimulationResult != "" {
		d, _, err := hash.DigestFile(cfg.SimulationResult)
		if err != nil {
			return types.Statement{}, err
		}
		predicate.SimulationResultDigest = d
	}

	subjects := make([]types.Subject, 0, 3)
	for _, p := range []string{cfg.RouteConfig, cfg.BudgetPolicy, cfg.FallbackGraph} {
		s, err := subjectFromPath(p)
		if err != nil {
			return types.Statement{}, err
		}
		subjects = append(subjects, s)
	}
	statement := newStatement(types.AttestationRoute, predicate, subjects, nil)
	setDependsOn(&statement, types.AttestationEval)
	return statement, nil
}
