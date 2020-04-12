use actix_files as fs;
use actix_web::{web, FromRequest};

use crate::handlers::{health, oauth, ui};
use crate::oauth::request::{TokenRequest, AuthorizationRequest};

pub fn routes(cfg: &mut web::ServiceConfig) {
    cfg.service(fs::Files::new("/static", "./dist").show_files_listing())
        .route("/health", web::get().to(health::get))
        .service(web::scope("/ui")
            .route("", web::get().to(ui::index)))
        .service(web::scope("/oauth")
            .configure(oauth::add_oauth_handler)
            .app_data(web::Query::<AuthorizationRequest>::configure(oauth::error::handle_query_errors))
            .app_data(web::Form::<TokenRequest>::configure(oauth::error::handle_form_errors))
            .route("/authorize", web::get().to(oauth::login_form))
            .route("/authorize", web::post().to(oauth::login))
            .route("/token", web::post().to(oauth::token)))
    ;
}
