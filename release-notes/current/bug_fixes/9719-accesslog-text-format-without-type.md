Fixed the File and ALS access log sinks silently falling back to the default JSON fields when `telemetry.accessLog.settings[].format` sets `text` without `type`, which the API accepts.
