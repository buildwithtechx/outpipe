use outpipe::{Client, CreateTunnel};
use std::{
    io::{Read, Write},
    net::{TcpListener, TcpStream},
    thread,
};

fn respond(mut stream: TcpStream, body: &str, status: &str) {
    let response = format!(
        "HTTP/1.1 {status}\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{body}",
        body.len()
    );
    stream
        .write_all(response.as_bytes())
        .expect("write response");
}

fn read_request(stream: &mut TcpStream) -> String {
    let mut request = Vec::new();
    let mut buffer = [0; 4096];
    loop {
        let count = stream.read(&mut buffer).expect("read request");
        if count == 0 {
            break;
        }
        request.extend_from_slice(&buffer[..count]);
        let Some(header_end) = request.windows(4).position(|window| window == b"\r\n\r\n") else {
            continue;
        };
        let headers = String::from_utf8_lossy(&request[..header_end]);
        let content_length = headers
            .lines()
            .find(|line| line.to_ascii_lowercase().starts_with("content-length:"))
            .and_then(|line| line.split_once(':'))
            .and_then(|(_, value)| value.trim().parse::<usize>().ok())
            .unwrap_or(0);
        if request.len() >= header_end + 4 + content_length {
            break;
        }
    }
    String::from_utf8(request).expect("request is valid HTTP text")
}

#[tokio::test]
async fn completes_tunnel_lifecycle_against_local_http_server() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind local server");
    let address = listener.local_addr().expect("local server address");
    let server = thread::spawn(move || {
        let expected = [
            ("POST", "/api/v1/organizations/org%2Fone/tunnels", "200 OK"),
            ("GET", "/api/v1/organizations/org%2Fone/tunnels", "200 OK"),
            ("GET", "/api/v1/tunnels/tunnel-1", "200 OK"),
            ("DELETE", "/api/v1/tunnels/tunnel-1", "204 No Content"),
            ("GET", "/api/v1/tunnels/error%2F401", "401 Unauthorized"),
        ];
        for (index, (method, path, status)) in expected.iter().enumerate() {
            let (mut stream, _) = listener.accept().expect("accept request");
            let request = read_request(&mut stream);
            let request_line = request.lines().next().expect("request line");
            assert!(
                request_line.starts_with(&format!("{method} {path} HTTP/1.1")),
                "{request_line}"
            );
            assert!(request.contains("authorization: Bearer integration-key"));
            if index == 0 {
                let body = request.split_once("\r\n\r\n").expect("request body").1;
                let payload: serde_json::Value =
                    serde_json::from_str(body).expect("valid JSON request body");
                assert_eq!(payload["targetHost"], "127.0.0.1");
                assert_eq!(payload["targetPort"], 3000);
            }
            let body = match index {
                0 | 2 => {
                    r#"{"id":"tunnel-1","name":"preview","protocol":"http","status":"active","targetHost":"127.0.0.1","targetPort":3000,"publicHostname":"preview.outpipe.app"}"#
                }
                1 => {
                    r#"[{"id":"tunnel-1","name":"preview","protocol":"http","status":"active","targetHost":"127.0.0.1","targetPort":3000}]"#
                }
                3 => "",
                _ => r#"{"message":"authentication required"}"#,
            };
            respond(stream, body, status);
        }
    });

    let client = Client::builder(format!("http://{address}"))
        .api_key("integration-key")
        .build()
        .expect("build client");
    let tunnel = CreateTunnel {
        name: "preview".into(),
        protocol: "http".into(),
        target_host: "127.0.0.1".into(),
        target_port: 3000,
        ..Default::default()
    };

    let created = client
        .create_tunnel("org/one", &tunnel)
        .await
        .expect("create tunnel");
    let listed = client.tunnels("org/one").await.expect("list tunnels");
    let inspected = client.tunnel("tunnel-1").await.expect("inspect tunnel");
    client
        .revoke_tunnel("tunnel-1")
        .await
        .expect("revoke tunnel");
    let error = client
        .tunnel("error/401")
        .await
        .expect_err("unauthorized request should fail");

    assert_eq!(created.id, "tunnel-1");
    assert_eq!(listed.len(), 1);
    assert_eq!(inspected.status, "active");
    assert!(matches!(
        error,
        outpipe::Error::Api {
            status: 401,
            message
        } if message == "authentication required"
    ));
    server.join().expect("local server completed");
}
