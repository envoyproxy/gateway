Removed the `--cpuset-threads` flag from the Envoy proxy startup arguments, which was causing Envoy to log a warning when the container's CPU limits did not align with a CPU set.
