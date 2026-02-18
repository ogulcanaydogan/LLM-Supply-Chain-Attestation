package llmsa.gates

import future.keywords.if
import future.keywords.in

default result := {"allow": false, "violations": ["rego result unavailable"]}

result := {"allow": false, "violations": ["Sensitive payload exposure blocked by policy."]} if {
  plaintext_blocked
}

result := {"allow": count(gate_violations) == 0, "violations": gate_violations} if {
  not plaintext_blocked
}

plaintext_blocked if {
  some s in input.statements
  s.privacy_mode == "plaintext_explicit"
  not allowed_plaintext[s.statement_id]
}

allowed_plaintext[id] if {
  some id in input.plaintext_allowlist
}

present_types[t] if {
  some s in input.statements
  t := s.attestation_type
}

gate_violations[msg] if {
  some g in input.gates
  gate_triggered(g.trigger_paths)
  missing := [r | some r in g.required_attestations; not present_types[r]]
  count(missing) > 0
  g.message != ""
  msg := g.message
}

gate_violations[msg] if {
  some g in input.gates
  gate_triggered(g.trigger_paths)
  missing := [r | some r in g.required_attestations; not present_types[r]]
  count(missing) > 0
  g.message == ""
  msg := sprintf("%s missing attestations: %v", [g.id, missing])
}

gate_triggered(patterns) if {
  some p in patterns
  some c in input.changed_files
  path_match(c, p)
}

path_match(path, pattern) if {
  endswith(pattern, "/**")
  prefix := trim_suffix(pattern, "/**")
  path == prefix
}

path_match(path, pattern) if {
  endswith(pattern, "/**")
  prefix := trim_suffix(pattern, "/**")
  startswith(path, sprintf("%s/", [prefix]))
}

path_match(path, pattern) if {
  glob.match(pattern, ["/"], path)
}
