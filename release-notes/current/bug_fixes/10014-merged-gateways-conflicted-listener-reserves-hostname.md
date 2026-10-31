Fixed merged Gateways where a listener already excluded by an earlier conflict-detection
pass (protocol or hostname conflict) could still reserve its `(protocol, hostname, port)`
slot in the cross-Gateway conflict check, wrongly marking a valid listener on a different
Gateway as conflicted and dropping it from the data plane.
