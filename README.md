[![Go build](https://github.com/Netcracker/qubership-core-lib-go-stomp-websocket/actions/workflows/go-build.yml/badge.svg)](https://github.com/Netcracker/qubership-core-lib-go-stomp-websocket/actions/workflows/go-build.yml)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?metric=coverage&project=Netcracker_qubership-core-lib-go-stomp-websocket)](https://sonarcloud.io/summary/overall?id=Netcracker_qubership-core-lib-go-stomp-websocket)
[![duplicated_lines_density](https://sonarcloud.io/api/project_badges/measure?metric=duplicated_lines_density&project=Netcracker_qubership-core-lib-go-stomp-websocket)](https://sonarcloud.io/summary/overall?id=Netcracker_qubership-core-lib-go-stomp-websocket)
[![vulnerabilities](https://sonarcloud.io/api/project_badges/measure?metric=vulnerabilities&project=Netcracker_qubership-core-lib-go-stomp-websocket)](https://sonarcloud.io/summary/overall?id=Netcracker_qubership-core-lib-go-stomp-websocket)
[![bugs](https://sonarcloud.io/api/project_badges/measure?metric=bugs&project=Netcracker_qubership-core-lib-go-stomp-websocket)](https://sonarcloud.io/summary/overall?id=Netcracker_qubership-core-lib-go-stomp-websocket)
[![code_smells](https://sonarcloud.io/api/project_badges/measure?metric=code_smells&project=Netcracker_qubership-core-lib-go-stomp-websocket)](https://sonarcloud.io/summary/overall?id=Netcracker_qubership-core-lib-go-stomp-websocket)

# go-stomp-websocket

Golang implementation of a STOMP client over WebSocket.

#### Supported operations:

* Establishing a STOMP connection
* Subscribing to events

#### Usage:

To start using the STOMP client:

1. Define the connection URL
2. Create the STOMP client using either a token or a custom Dial
3. Define a channel to receive frames

#### Example

To connect to the Watch API of Tenant-Manager, which uses the STOMP protocol and has the following URL:

```
ws://tenant-manager:8080/api/v3/tenant-manager/watch
```

#### Creating the STOMP client

##### Using a token

```go
token, _ := tenantWatchClient.Credential.GetAuthToken()
url, _ := url.Parse("ws://localhost:8080/api/v3/tenant-manager/watch")
stompClient, _ := go_stomp_websocket.ConnectWithToken(*url, token)
```

##### Using a custom Dial

```go
type ConnectionDialer interface {
    Dial(webSocketURL url.URL, dialer websocket.Dialer, requestHeaders http.Header) (*websocket.Conn, *http.Response, error)
}
```

```go
url, _ := url.Parse("ws://localhost:8080/api/v3/tenant-manager/watch")
dialer := websocket.Dialer{}
// configure the dialer
requestHeaders := http.Header{}
// add headers
connDial := ConnectionDialerImpl{} // implements the Dial method of ConnectionDialer interface
stompClient, _ := go_stomp_websocket.Connect(*url, dialer, requestHeaders, connDial)
```

Subscribe to events:

```go
subscr, _ := stompClient.Subscribe("/tenant-changed")
```

Handle received frames:

```go
go func() {
    for {
        var tenant = new(tenant.Tenant)
        frame := <-subscr.FrameCh // Receive frame
        if len(frame.Body) > 0 {
            err := json.Unmarshal([]byte(frame.Body), tenant) // Parse the frame body into Tenant structure
            if err != nil {
                fmt.Println(err)
            } else {
                fmt.Printf("Received tenant with id:%s\n", tenant.ObjectId)
            }
        }
    }
}()
```

