## Message Service Design

### Goals
* Outline the design for the Message Service, a mandatory component in the [System Design](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md).

### Non-Goals
* Add implementation details of the Message Service.

### Motivation
The Message service described [here](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md#message-service) allows for internal
service components within the Envoy Gateway project to
* Be loosely coupled.
* Run asynchronously.
* Reside in multiple processes and rely on the Message Service to transport the messages accurately.
* Have multiple publishers e.g. if Envoy Gateway allows multiple [Config Sources](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md#config-sources) to be enabled at the same time, the Message Service can act as an intermediate entity to aggregate resources before sending
it to the subscribers.
* Have multiple subscribers e.g. the Message Service enables multiple subscribers such as the [xDS Translator](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md#xds-translator) and the [Provisioner](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md#provisioner) to subscribe to [IR](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md#intermediate-representation-ir) resources published by the [Gateway API Translator](https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md#gateway-api-translator).

### API

#### Instantiating an In-Memory Message Service
```
msgSvc  = message.NewMemoryService()
```

#### Assigning a Publisher handle to a Service to publish a specific message type
```
svc.Publisher = msgSvc.Publisher([]message.MessageType{message.GatewayClassMessageType})
```

#### Publishing a Message
```
msg := message.Message{
  ID:     <id>,
  Type:   message.GatewayClassMessageType,
  Source: <source>,
  Value: <value>,
  Timestamp: <timestamp>,
}

svc.Publisher.Publish(msg)
```

#### Assigning a Subscriber handle to a Service allowing it to subscribe to a specific message type
```
svc.Subscriber = msgSvc.Subscriber([]message.MessageType{message.GatewayClassMessageType})
```

#### Service Subscribing to Messages by calling Subscribe (a blocking function)
```
svc.Subscriber.Subscribe(svc.handleFunc)
```

#### Receiving and handling Message Types
```
func (svc *Service) handleFunc(v interface{}) error {
	switch payload := m.Value.(type) {
	case message.GatewayClassMessageType:
		// do something
	}
	return nil
}
```

### Notes
* The Message Service does not dictate whether the messages sent between
the publisher and subscriber are incremental or represent SoTW (State of
the World).
* The Message Service API can be extended in the future so it can store and aggregate
messages from all publishers before sending them to the subscribers.
* The `Subscribe` API calls `handleFunc` sequentially. 


