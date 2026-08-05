Fixed EnvoyGateway config hot-reload to apply defaults before validation, so validators always run against a fully-defaulted struct on both the startup and reload paths.
