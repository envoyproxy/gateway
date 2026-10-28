Added `maxConnectionDurationJitter` to the connection limit settings in ClientTrafficPolicy, so connections opened at the same time do not all reach their max duration and reconnect together.
