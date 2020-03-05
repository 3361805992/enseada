use actix_web::middleware::Logger;
use actix_web::{App, HttpServer};

use crate::couchdb::add_couch_client;
use crate::routes::routes;

pub async fn run() -> std::io::Result<()> {
    HttpServer::new(move || {
        App::new()
            .wrap(Logger::default())
            .configure(add_couch_client)
            .configure(routes)
    })
    .bind("0.0.0.0:9623")?
    .run()
    .await
}
