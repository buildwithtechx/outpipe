use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
pub struct Organization {
    pub id: String,
    pub name: String,
    pub slug: String,
    #[serde(rename = "ownerId")]
    pub owner_id: Option<String>,
}

#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
pub struct Tunnel {
    pub id: String,
    pub name: String,
    pub protocol: String,
    pub status: String,
    #[serde(rename = "publicHostname")]
    pub public_hostname: Option<String>,
    #[serde(rename = "publicPort")]
    pub public_port: Option<u16>,
    #[serde(rename = "targetHost")]
    pub target_host: String,
    #[serde(rename = "targetPort")]
    pub target_port: u16,
}

#[derive(Debug, Clone, Deserialize, Serialize, Default, PartialEq)]
pub struct CreateTunnel {
    pub name: String,
    pub protocol: String,
    #[serde(rename = "targetHost")]
    pub target_host: String,
    #[serde(rename = "targetPort")]
    pub target_port: u16,
    #[serde(rename = "publicHostname", skip_serializing_if = "Option::is_none")]
    pub public_hostname: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub password: Option<String>,
}
