use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("invalid client configuration: {0}")]
    Configuration(String),
    #[error("request failed: {0}")]
    Request(#[from] reqwest::Error),
    #[error("invalid URL: {0}")]
    Url(#[from] url::ParseError),
    #[error("API request failed with status {status}: {message}")]
    Api { status: u16, message: String },
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),
    #[error("protocol error: {0}")]
    Protocol(String),
}

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ApiError {
    pub status: u16,
    pub message: String,
}
