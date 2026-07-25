Fixed a data race between config-reload and the standalone server's shutdown/error logging by having the config loader publish logger updates over a channel instead of a shared, unguarded field.
