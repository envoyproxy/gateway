Fixed ListenerSet not being reconciled when a referenced TLS Secret is
created or updated after the ListenerSet. Secret watches previously only
indexed Gateway certificateRefs, so cert-manager style late Secret creation
left the ListenerSet stuck with Programmed=False until an unrelated reconcile.
