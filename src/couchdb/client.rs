use derivative::Derivative;
use reqwest::{Client as HttpClient, Method, StatusCode, Response};
use serde::de::DeserializeOwned;
use serde::ser::Serialize;
use url::{ParseError, Url};
use actix_web::web::method;
use crate::couchdb::responses::Ok;

#[derive(Derivative)]
#[derivative(Debug, Clone)]
pub(super) struct Client {
    client: HttpClient,
    base_url: Url,
    username: String,
    #[derivative(Debug = "ignore")]
    password: Option<String>,
}

impl Client {
    pub fn new(base_url: Url, username: String, password: String) -> Client {
        let client = HttpClient::builder()
            .use_rustls_tls()
            .build()
            .expect("HttpClient::build()");
        Client {
            client,
            base_url,
            username,
            password: Some(password),
        }
    }

    pub async fn get<T: DeserializeOwned>(&self, path: &str) -> reqwest::Result<T> {
        self.request(Method::GET, path, None::<bool>, None::<bool>).await
    }

    pub async fn put<B: Serialize, Q: Serialize, R: DeserializeOwned>(
        &self,
        path: &str,
        body: Option<B>,
        query: Option<Q>
    ) -> reqwest::Result<R> {
        self.request(Method::PUT, path, body, query).await
    }

    pub async fn post<B: Serialize, Q: Serialize, R: DeserializeOwned>(
        &self,
        path: &str,
        body: Option<B>,
        query: Option<Q>
    ) -> reqwest::Result<R> {
        self.request(Method::POST, path, body, query).await
    }

    pub async fn delete<Q: Serialize>(&self, path: &str, query: Option<Q>) -> reqwest::Result<()> {
        self.request(Method::DELETE, path, None::<bool>, query).await.map(|_: Ok| ())
    }

    pub async fn exists(&self, path: &str) -> reqwest::Result<bool> {
        let result = self
            .client
            .head(self.build_url(path).unwrap())
            .basic_auth(&self.username, self.password.as_ref())
            .send()
            .await?
            .error_for_status();

        match result {
            Ok(_res) => Ok(true),
            Err(err) => match err.status() {
                Some(StatusCode::NOT_FOUND) => Ok(false),
                _ => Err(err),
            },
        }
    }

    pub(crate) fn build_url(&self, path: &str) -> Result<Url, ParseError> {
        self.base_url.join(path)
    }

    pub async fn request<B: Serialize, Q: Serialize, R: DeserializeOwned>(
        &self,
        method: Method,
        path: &str,
        body: Option<B>,
        query: Option<Q>
    ) -> reqwest::Result<R> {
        let req = self
            .client
            .request(method, self.build_url(path).unwrap())
            .basic_auth(&self.username, self.password.as_ref());
        let req = if let Some(body) = body {
            req.json::<B>(&body)
        } else {
            req
        };

        let req = if let Some(query) = query {
            req.query::<Q>(&query)
        } else {
            req
        };

        req.send().await?.error_for_status()?.json().await
    }
}
