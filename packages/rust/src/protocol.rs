use serde::{de::DeserializeOwned, Deserialize, Serialize};

pub const VERSION: u8 = 1;
pub const MAX_FRAME_SIZE: usize = 32 * 1024 * 1024;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Envelope<T = serde_json::Value> {
    pub version: u8,
    #[serde(rename = "type")]
    pub message_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub request_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub payload: Option<T>,
}

pub fn encode<T: Serialize>(message: &Envelope<T>) -> crate::Result<Vec<u8>> {
    if message.version != VERSION {
        return Err(crate::Error::Protocol(
            "unsupported protocol version".into(),
        ));
    }
    if message.message_type.is_empty() {
        return Err(crate::Error::Protocol("message type is required".into()));
    }
    let data = serde_json::to_vec(message)?;
    if data.len() > MAX_FRAME_SIZE {
        return Err(crate::Error::Protocol("frame exceeds maximum size".into()));
    }
    Ok(data)
}

pub fn decode<T: DeserializeOwned>(data: &[u8]) -> crate::Result<Envelope<T>> {
    if data.len() > MAX_FRAME_SIZE {
        return Err(crate::Error::Protocol("frame exceeds maximum size".into()));
    }
    let message: Envelope<T> = serde_json::from_slice(data)?;
    if message.version != VERSION || message.message_type.is_empty() {
        return Err(crate::Error::Protocol("invalid protocol envelope".into()));
    }
    Ok(message)
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct AuthRequest {
    pub token: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub requested_capabilities: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct VersionNegotiate {
    pub min_version: u8,
    pub max_version: u8,
    pub client_name: String,
    pub client_version: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct OpenTunnel {
    pub token: String,
    pub local_port: u16,
    pub protocol: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tunnel_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub subdomain: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub custom_domain: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct OpenTunnelAck {
    pub tunnel_id: String,
    pub public_url: String,
    pub public_port: Option<u16>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct FlowControl {
    pub stream_id: String,
    pub action: String,
    pub window_size: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Data {
    pub tunnel_id: String,
    pub stream_id: String,
    pub data: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct HttpRequest {
    pub method: String,
    pub path: String,
    pub headers: std::collections::HashMap<String, Vec<String>>,
    pub body: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct HttpResponse {
    pub status_code: u16,
    pub headers: std::collections::HashMap<String, Vec<String>>,
    pub body: Option<String>,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct TcpData {
    pub tunnel_id: Option<String>,
    pub connection_id: String,
    pub data: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct TcpClose {
    pub tunnel_id: Option<String>,
    pub connection_id: String,
    pub reason: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct UdpData {
    pub tunnel_id: Option<String>,
    pub packet_id: String,
    pub source_address: String,
    pub source_port: u16,
    pub data: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct UdpResponse {
    pub tunnel_id: Option<String>,
    pub packet_id: String,
    pub target_address: String,
    pub target_port: u16,
    pub data: String,
}
