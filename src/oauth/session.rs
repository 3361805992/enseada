use serde::{Serialize, Deserialize};
use crate::oauth::scope::Scope;

#[derive(Clone, Debug, Default, Deserialize, Serialize)]
pub struct Session {
    client_id: String,
    scope: Scope,
    user_id: Option<String>,
}

impl Session {
    pub fn empty() -> Session {
        Default::default()
    }

    pub fn client_id(&self) -> &String {
        &self.client_id
    }

    pub fn set_client_id(&mut self, client_id: String) -> &mut Self {
        self.client_id = client_id;
        self
    }

    pub fn scope(&self) -> &Scope {
        &self.scope
    }

    pub fn set_scope(&mut self, scope: Scope) -> &mut Self {
        self.scope = scope;
        self
    }

    pub fn user_id(&self) -> &Option<String> {
        &self.user_id
    }

    pub fn set_user_id(&mut self, user_id: String) -> &mut Self {
        self.user_id = Some(user_id);
        self
    }
}