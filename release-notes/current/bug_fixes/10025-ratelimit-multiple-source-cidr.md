Fixed a global rate limit rule with more than one `sourceCIDR` across its
`clientSelectors` silently keeping only the last one. The earlier selectors were
dropped during translation with no error and the policy still reported
`Accepted: True`, so a rule written as "limit each client individually, except
this CIDR" enforced only half of what it said. Such a rule is now rejected with
an `Accepted: False` condition explaining that only one `clientSelector` per rule
may specify `sourceCIDR`.
