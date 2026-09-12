Fixed a stale `status.parents` entry lingering on a Route after one of its `spec.parentRefs` was removed, while still keeping status entries for parents not touched in a given reconciliation batch.
