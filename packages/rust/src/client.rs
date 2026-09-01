use crate::{
    error::Error,
    models::{CreateTunnel, Organization, Tunnel},
    Result,
};
use reqwest::{Client as HttpClient, Method, Response};
use serde::de::DeserializeOwned;
use serde::Serialize;
use url::Url;

pub struct ClientBuilder {
    base_url: String,
    api_key: Option<String>,
    http: Option<HttpClient>,
}

pub struct Client {
    base_url: Url,
    api_key: Option<String>,
    http: HttpClient,
}

impl ClientBuilder {
    pub fn new(base_url: impl Into<String>) -> Self {
        Self {
            base_url: base_url.into(),
            api_key: None,
            http: None,
        }
    }

    pub fn api_key(mut self, api_key: impl Into<String>) -> Self {
        self.api_key = Some(api_key.into());
        self
    }

    pub fn http_client(mut self, http: HttpClient) -> Self {
        self.http = Some(http);
        self
    }

    pub fn build(self) -> Result<Client> {
        let mut base_url = Url::parse(&self.base_url)?;
        if !base_url.path().ends_with('/') {
            base_url.set_path(&format!("{}/", base_url.path()));
        }
        if base_url.scheme() != "http" && base_url.scheme() != "https" {
            return Err(Error::Configuration(
                "base URL must use HTTP or HTTPS".into(),
            ));
        }
        Ok(Client {
            base_url,
            api_key: self.api_key,
            http: self.http.unwrap_or_default(),
        })
    }
}

impl Client {
    pub fn builder(base_url: impl Into<String>) -> ClientBuilder {
        ClientBuilder::new(base_url)
    }

    pub async fn health(&self) -> Result<serde_json::Value> {
        self.get("healthz", None::<&()>).await
    }

    pub async fn ready(&self) -> Result<serde_json::Value> {
        self.get("readyz", None::<&()>).await
    }

    pub async fn organizations(&self) -> Result<Vec<Organization>> {
        self.get("api/v1/organizations", None::<&()>).await
    }

    pub async fn organization(&self, id: &str) -> Result<Organization> {
        self.get(&format!("api/v1/organizations/{}", encode(id)), None::<&()>)
            .await
    }

    pub async fn tunnels(&self, organization_id: &str) -> Result<Vec<Tunnel>> {
        self.get(
            &format!("api/v1/organizations/{}/tunnels", encode(organization_id)),
            None::<&()>,
        )
        .await
    }

    pub async fn create_tunnel(
        &self,
        organization_id: &str,
        tunnel: &CreateTunnel,
    ) -> Result<Tunnel> {
        self.send(
            Method::POST,
            &format!("api/v1/organizations/{}/tunnels", encode(organization_id)),
            Some(tunnel),
            None::<&()>,
        )
        .await
    }

    pub async fn tunnel(&self, tunnel_id: &str) -> Result<Tunnel> {
        self.get(
            &format!("api/v1/tunnels/{}", encode(tunnel_id)),
            None::<&()>,
        )
        .await
    }

    pub async fn set_tunnel_status(&self, tunnel_id: &str, status: &str) -> Result<()> {
        self.send_empty(
            Method::PATCH,
            &format!("api/v1/tunnels/{}/status", encode(tunnel_id)),
            &serde_json::json!({"status": status}),
        )
        .await
    }

    pub async fn revoke_tunnel(&self, tunnel_id: &str) -> Result<()> {
        self.send_empty(
            Method::DELETE,
            &format!("api/v1/tunnels/{}", encode(tunnel_id)),
            &(),
        )
        .await
    }

    async fn get<T: DeserializeOwned, Q: Serialize>(
        &self,
        path: &str,
        query: Option<Q>,
    ) -> Result<T> {
        self.send(Method::GET, path, None::<&()>, query).await
    }

    async fn send<T: DeserializeOwned, B: Serialize, Q: Serialize>(
        &self,
        method: Method,
        path: &str,
        body: Option<B>,
        query: Option<Q>,
    ) -> Result<T> {
        let response = self.request(method, path, body, query).await?;
        if response.status().as_u16() == 204 {
            return serde_json::from_value(serde_json::Value::Null).map_err(Error::from);
        }
        Ok(response.json().await?)
    }

    async fn send_empty<B: Serialize>(&self, method: Method, path: &str, body: &B) -> Result<()> {
        self.request(method, path, Some(body), None::<&()>)
            .await?
            .error_for_status()?;
        Ok(())
    }

    async fn request<B: Serialize, Q: Serialize>(
        &self,
        method: Method,
        path: &str,
        body: Option<B>,
        query: Option<Q>,
    ) -> Result<Response> {
        let mut request = self.http.request(method, self.base_url.join(path)?);
        if let Some(key) = &self.api_key {
            request = request.bearer_auth(key);
        }
        if let Some(query) = query {
            request = request.query(&query);
        }
        if let Some(body) = body {
            request = request.json(&body);
        }
        let response = request.send().await?;
        if !response.status().is_success() {
            let status = response.status().as_u16();
            let message = response
                .text()
                .await
                .unwrap_or_else(|_| "request failed".into());
            return Err(Error::Api { status, message });
        }
        Ok(response)
    }
}

fn encode(value: &str) -> String {
    url::form_urlencoded::byte_serialize(value.as_bytes()).collect()
}
