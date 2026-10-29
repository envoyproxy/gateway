Fixed status updates being dropped when the API server returned a timeout, throttling or another transient error. The status writer now retries these errors with backoff, in addition to conflicts.
