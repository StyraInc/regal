package p

default allow = false

allow if {
    isEmployee
    "developer" in input.user.roles
}

isEmployee if regex.match("@acmecorp\\.com$", input.user.email)
