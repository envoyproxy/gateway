Roadmap
------------

This document serves as a high-level reference for Envoy Gateway users
and contributors to understand the direction of the project.

Contributing to the Roadmap
~~~~~~~~~~~~~~~~~~~~~~~~~~~

-  To add a feature to the roadmap, create an
   `issue <https://github.com/envoyproxy/gateway/issues>`__ or join a
   `community
   meeting <https://docs.google.com/document/d/1leqwsHX8N-XxNEyTflYjRur462ukFxd19Rnk3Uzy55I/edit?usp=sharing>`__
   to discuss your use case. If your feature is accepted, a maintainer
   will assign your issue to a `release
   milestone <https://github.com/envoyproxy/gateway/milestones>`__ and
   update this document accordingly.
-  To help with an existing roadmap item, comment on or assign yourself
   to the associated issue.
-  If a roadmap item doesn’t have an issue, create one, assign yourself
   to the issue, and reference this document. A maintainer will submit a
   `pull request <https://github.com/envoyproxy/gateway/compare>`__ to
   add the feature to the roadmap. **Note:** The feature should be
   discussed in an issue or a community meeting before implementing it.

If you don’t know where to start contributing, help is needed to reduce
technical, automation, and documentation debt. Look for issues with the
``help wanted`` label to get started.

Details
~~~~~~~

Roadmap features and timelines may change based on feedback, community
contributions, etc. If you depend on a specific roadmap item, you’re
encouraged to attend a community meeting to discuss the details, or help
us deliver the feature by contributing to the project.

``Last Updated: October 2022``

.. _v0.2.0v0.2.0-establish-a-solid-foundation:

`v0.2.0 <https://github.com/envoyproxy/gateway/milestone/1>`__: Establish a Solid Foundation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-  Complete the core Envoy Gateway implementation- `Issue
   #60 <https://github.com/envoyproxy/gateway/issues/60>`__.
-  Establish initial testing, e2e, integration, etc- `Issue
   #64 <https://github.com/envoyproxy/gateway/issues/64>`__.
-  Establish user and developer project documentation- `Issue
   #17 <https://github.com/envoyproxy/gateway/issues/17>`__.
-  Achieve Gateway API conformance (e.g. routing, LB, Header
   transformation, etc.)- `Issue
   #65 <https://github.com/envoyproxy/gateway/issues/65>`__.
-  Setup a CI/CD pipeline- `Issue
   #63 <https://github.com/envoyproxy/gateway/issues/63>`__.

.. _v0.3.0v0.3.0-drive-advanced-features-through-extension-mechanisms:

`v0.3.0 <https://github.com/envoyproxy/gateway/milestone/7>`__: Drive Advanced Features through Extension Mechanisms
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-  Global Rate Limiting
-  AuthN/AuthZ- `Issue
   #336 <https://github.com/envoyproxy/gateway/issues/336>`__.
-  Lets Encrypt Integration

.. _v0.4.0v0.4.0-manageability-and-scale:

`v0.4.0 <https://github.com/envoyproxy/gateway/milestone/12>`__: Manageability and Scale
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-  Tooling for devs/infra admins to aid in managing/maintaining EG
-  Support advanced provisioning use cases (e.g. multi-cluster,
   serverless, etc.)
-  Perf testing (EG specifically)
-  Support for Chaos engineering?
