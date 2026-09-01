# outpipe

Async Rust client and protocol types for Outpipe.

```toml
[dependencies]
outpipe = "0.1"
```

```rust
use outpipe::{Client, CreateTunnel};

let client = Client::builder("https://api.outpipe.dev")
    .api_key(std::env::var("OUTPIPE_API_KEY")?)
    .build()?;

let tunnel = client.create_tunnel("organization-id", &CreateTunnel {
    name: "checkout-preview".into(),
    protocol: "https".into(),
    target_host: "127.0.0.1".into(),
    target_port: 3000,
    ..Default::default()
}).await?;
```

The crate exposes the versioned relay envelope and core authentication/open
tunnel payloads through `outpipe::protocol`.
