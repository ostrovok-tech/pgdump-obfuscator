package main

type Configuration struct {
	Obfuscations []TargetedObfuscation
}

// TODO: read from file?
var Config *Configuration = &Configuration{
	Obfuscations: []TargetedObfuscation{
		TargetedObfuscation{
			Target{Table: "auth_user", Column: "email"},
			ScrambleEmail,
		},
		TargetedObfuscation{
			Target{Table: "auth_user", Column: "password"},
			ScrambleBytes,
		},
		TargetedObfuscation{
			Target{Table: "accounts_profile", Column: "phone"},
			ScramblePhone,
		},
	},
}
