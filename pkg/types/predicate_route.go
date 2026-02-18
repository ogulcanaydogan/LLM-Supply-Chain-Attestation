package types

type ProviderModel struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type RoutePredicate struct {
	RouteConfigDigest      string          `json:"route_config_digest"`
	ProviderSet            []ProviderModel `json:"provider_set"`
	BudgetPolicyDigest     string          `json:"budget_policy_digest"`
	FallbackGraphDigest    string          `json:"fallback_graph_digest"`
	RoutingStrategy        string          `json:"routing_strategy"`
	CanaryConfigDigest     string          `json:"canary_config_digest,omitempty"`
	SimulationResultDigest string          `json:"simulation_result_digest,omitempty"`
}
