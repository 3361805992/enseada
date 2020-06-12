use actix_files as fs;
use actix_web::{web, FromRequest};

use crate::http::handler::{api_docs, health, home, oauth, ui};
use crate::oauth::request::{AuthorizationRequest, TokenRequest};

pub fn routes(cfg: &mut web::ServiceConfig) {
    cfg.route("/", web::get().to(home))
        .service(fs::Files::new("/static", "./dist"))
        .service(fs::Files::new("/images", "./images"))
        .route("/health", web::get().to(health::get))
        // UI
        .service(web::scope("/ui").route("", web::get().to(ui::index)))
        // OAuth
        .service(
            web::scope("/oauth")
                .app_data(web::Query::<AuthorizationRequest>::configure(
                    oauth::error::handle_query_errors,
                ))
                .app_data(web::Form::<TokenRequest>::configure(
                    oauth::error::handle_form_errors,
                ))
                .service(
                    web::resource("/authorize")
                        .route(web::get().to(oauth::login_form))
                        .route(web::post().to(oauth::login)),
                )
                .route("/token", web::post().to(oauth::token))
                .route("/introspect", web::post().to(oauth::introspect))
                .route("/revoke", web::post().to(oauth::revoke)),
        )
        // API
        .service(
            web::scope("/api").service(
                web::scope("/docs")
                    .service(
                        web::resource("/openapi.yml")
                            .name("open_api_spec")
                            .route(web::get().to(api_docs::open_api)),
                    )
                    .route("", web::get().to(api_docs::redoc)),
            ),
        );
}
