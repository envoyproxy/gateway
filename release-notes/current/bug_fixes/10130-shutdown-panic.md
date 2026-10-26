Fixed a panic during shutdown caused by closing the shared `infraIR`/`pResources` channels and resources maps before all runner goroutines still consuming or writing to them had exited.
