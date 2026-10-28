Backend now requires `endpoints` when `type` is `Endpoints` (the default). Such Backends previously passed API validation even though they had no destination to route to, and are now rejected.
