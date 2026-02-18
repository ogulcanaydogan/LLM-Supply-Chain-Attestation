package llmsa.gates

default allow = false

allow {
  count(input.violations) == 0
}
