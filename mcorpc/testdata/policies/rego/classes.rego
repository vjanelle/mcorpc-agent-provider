package choria.mcorpc.authpolicy

default allow = false

allow {
    # Only allow a matching list
    input.classes = ["alpha", "beta"]
    # Only allow if classes is defined
    input.classes[_] = "alpha"
    input.classes[_] = "beta"
}
