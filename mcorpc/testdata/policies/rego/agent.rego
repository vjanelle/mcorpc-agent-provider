package choria.mcorpc.authpolicy

default allow = false

allow {
    # Only allow a matching list
    input.agents = ["stub_agent", "buts_agent"]
    # Only allow if classes is defined
    input.agents[_] = "stub_agent"
    input.agents[_] = "buts_agent"
}
