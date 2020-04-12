use serde::{Serialize, Deserialize};

use std::any::Any;
use std::iter::Map;

#[derive(Deserialize, Serialize, Debug)]
pub struct DBInfo {
    pub cluster: DBClusterInfo,
    pub compact_running: bool,
    pub db_name: String,
    pub disk_format_version: i32,
    pub doc_count: i64,
    pub doc_del_count: i64,
    pub instance_start_time: String,
    pub purge_seq: String,
    pub sizes: DBSizes,
    pub update_seq: String,
    pub props: DBProps,
}

#[derive(Deserialize, Serialize, Debug)]
pub struct DBSizes {
    active: i32,
    external: i32,
    file: i32,
}

#[derive(Deserialize, Serialize, Debug)]
pub struct DBClusterInfo {
    pub n: i16,
    pub q: i16,
    pub r: i16,
    pub w: i16,
}

#[derive(Deserialize, Serialize, Debug)]
pub struct DBProps {
    pub partitioned: Option<bool>,
}

#[derive(Deserialize, Serialize, Debug)]
pub struct Ok {
    pub ok: bool,
}
