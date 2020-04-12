use io::BufReader;
use std::fs::File;
use std::io;
use std::io::{Seek, SeekFrom};

use actix_web::{App, HttpServer};
use actix_web::middleware::{DefaultHeaders, Logger};
use rustls::{Certificate, NoClientAuth, PrivateKey, ServerConfig};
use rustls::internal::pemfile::{certs, pkcs8_private_keys, rsa_private_keys};

use crate::config::CONFIG;
use crate::couchdb::add_couch_client;
use crate::routes::routes;

pub async fn run() -> io::Result<()> {
    let address = format!("0.0.0.0:{}", CONFIG.port());
    let public_host = CONFIG.public_host();
    let tls = CONFIG.tls();

    let server = HttpServer::new(move || {
        App::new()
            .wrap(Logger::default())
            .wrap(default_headers())
            .configure(add_couch_client)
            .configure(routes)
    });

    let server = if let Some(public_host) = public_host {
        server.server_hostname(public_host)
    } else {
        server
    };

    let server = if tls.enabled() {
        let mut config = ServerConfig::new(NoClientAuth::new());
        let cert_f = &mut File::open(tls.cert_path().expect("missing tls.cert_path"))?;
        let key_f = &mut File::open(tls.key_path().expect("missing tls.cert_path"))?;
        let certs = get_certs(cert_f);
        let key = get_rsa_key(key_f);
        config.set_single_cert(certs, key).unwrap();
        server.bind_rustls(&address, config)
    } else {
        server.bind(&address)
    }?;

    log::info!("Server started listening on {}", &address);
    server.run().await
}

fn get_certs(cert: &File) -> Vec<Certificate> {
    let buf = &mut BufReader::new(cert);
    certs(buf).unwrap()
}

fn get_rsa_key(key: &mut File) -> PrivateKey {
    let rsa_buf = &mut BufReader::new(key.try_clone().unwrap());
    let pkcs_buf = &mut BufReader::new(key.try_clone().unwrap());
    let rsa = rsa_private_keys(rsa_buf).unwrap();
    key.seek(SeekFrom::Start(0)).unwrap();
    let pkcs8 = pkcs8_private_keys(pkcs_buf).unwrap();
    rsa.first().or(pkcs8.first())
        .expect("key format not supported. must be either RSA or PKCS8-encoded.")
        .clone()
}

fn default_headers() -> DefaultHeaders{
    let h = DefaultHeaders::new();
    if CONFIG.tls().enabled() {
        h.header("Strict-Transport-Security", "max-age=31536000")
    } else {
        h
    }
}
