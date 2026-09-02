# Outpipe Go SDK

Reusable Go HTTP and relay client for Outpipe.

```go
client, err := client.New(client.Config{
    BaseURL: "https://api.outpipe.dev",
    APIKey: os.Getenv("OUTPIPE_API_KEY"),
})
tunnels, err := client.Tunnels(ctx, organizationID)
```

The `protocol` package contains the versioned relay envelope and payload
types. `client.OpenRelay` performs relay negotiation, authentication, and
tunnel opening for native workers.
