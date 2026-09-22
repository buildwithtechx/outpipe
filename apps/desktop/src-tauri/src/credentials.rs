use keyring::Entry;
use serde::{Deserialize, Serialize};

const SERVICE: &str = "outpipe.desktop";

#[derive(Serialize, Deserialize, Clone, Copy)]
#[serde(rename_all = "snake_case")]
pub enum CredentialKind {
    ApiKey,
    AgentToken,
}

impl CredentialKind {
    fn account(self) -> &'static str {
        match self {
            CredentialKind::ApiKey => "api_key",
            CredentialKind::AgentToken => "agent_token",
        }
    }
}

#[derive(Serialize)]
pub struct CredentialsStatus {
    pub has_api_key: bool,
    pub has_agent_token: bool,
}

fn entry(kind: CredentialKind) -> Result<Entry, String> {
    Entry::new(SERVICE, kind.account()).map_err(|error| format!("open keyring: {error}"))
}

#[tauri::command]
pub fn credentials_save(kind: CredentialKind, secret: String) -> Result<(), String> {
    if secret.trim().is_empty() {
        return Err("credential cannot be empty".to_string());
    }
    entry(kind)?
        .set_password(&secret)
        .map_err(|error| format!("store credential: {error}"))
}

#[tauri::command]
pub fn credentials_clear(kind: CredentialKind) -> Result<(), String> {
    let entry = entry(kind)?;
    match entry.delete_credential() {
        Ok(()) => Ok(()),
        Err(keyring::Error::NoEntry) => Ok(()),
        Err(error) => Err(format!("clear credential: {error}")),
    }
}

#[tauri::command]
pub fn credentials_status() -> Result<CredentialsStatus, String> {
    Ok(CredentialsStatus {
        has_api_key: credential(CredentialKind::ApiKey).unwrap_or(None).is_some(),
        has_agent_token: credential(CredentialKind::AgentToken).unwrap_or(None).is_some(),
    })
}

pub(crate) fn credential(kind: CredentialKind) -> Result<Option<String>, String> {
    let Ok(entry) = entry(kind) else {
        return Ok(None);
    };
    match entry.get_password() {
        Ok(secret) => Ok(Some(secret)),
        Err(keyring::Error::NoEntry) => Ok(None),
        Err(_) => Ok(None),
    }
}
