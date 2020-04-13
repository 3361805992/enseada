use std::pin::Pin;

use actix_web::{Error, FromRequest, HttpRequest};
use actix_web::dev::{Payload, PayloadStream};
use actix_web::error::PayloadError;
use actix_web::web::{Bytes, Data};
use futures::{Future, FutureExt, Stream, TryFutureExt};

use crate::http::error::ApiError;
use crate::http::extractor::session::TokenSession;
use crate::oauth::scope::Scope as OAuthScope;
use crate::oauth::session::Session;

pub type Scope = OAuthScope;

impl FromRequest for Scope {
    type Error = ApiError;
    type Future = Pin<Box<dyn Future<Output=Result<Self, Self::Error>>>>;
    type Config = ();

    fn from_request(req: &HttpRequest, payload: &mut Payload<PayloadStream>) -> Self::Future {
        let session_fut = TokenSession::from_request(req, payload);
        Box::pin(async move {
            let session: Session = session_fut.await?;
            Ok(session.scope().clone())
        })
    }
}