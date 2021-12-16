# Prototype of prometheus event receiver that can create k8s events

Idea is that you would create an AlertManagerConfig per CR with the URL including query parameters that allow you to identify the `InvolvedObject`, and it emits an event based on that, which should trigger reconciliation on anything watching that resource.


Usage:
```
go run main.go
```

to kick off the APIServer, after which you should be able to just POST prometheus alert requests at it.

An example request payload is included in the `request.json` file, to use that you would run:

```
curl -XPOST 'localhost:6000?kind=Hello&group=apps.example.com&version=v1beta1&name=test&namespace=test' -d @request.json
```
