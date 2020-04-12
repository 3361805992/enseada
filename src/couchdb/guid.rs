use serde::{Deserialize, Serialize, Serializer, Deserializer};

pub struct Guid {
    partition: Option<String>,
    id: String,
}

impl Guid {
    pub fn partition(&self) -> Option<String> {
        self.partition.clone()
    }

    pub fn id(&self) -> &String {
        &self.id
    }
}

impl ToString for Guid {
    fn to_string(&self) -> String {
        let partition = self.partition().map(|s| format!("{}:", s.clone()))
            .unwrap_or("".to_string());
        format!("{}{}", partition, &self.id)
    }
}

impl From<String> for Guid {
    fn from(s: String) -> Self {
        if s.contains(":") {
            let p: Vec<&str> = s.splitn(2, ":").collect();
            let partition = p.get(0)
                .take()
                .cloned()
                .map(String::from);
            let id = p.get(1)
                .take()
                .cloned()
                .map(String::from)
                .unwrap();
            Guid { partition, id }
        } else {
            Guid { partition: None, id: s }
        }
    }
}

impl<'a> From<&'a str> for Guid {
    fn from(s: &'a str) -> Self {
        Guid::from(s.to_string())
    }
}

impl Serialize for Guid {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error> where
        S: Serializer {
        self.to_string().serialize(serializer)
    }
}

impl<'de> Deserialize<'de> for Guid {
    fn deserialize<D>(deserializer: D) -> Result<Self, <D as Deserializer<'de>>::Error> where
        D: Deserializer<'de> {
        let s = String::deserialize(deserializer)?;
        Ok(Guid::from(s))
    }
}

#[cfg(test)]
mod test {
    use crate::couchdb::guid::Guid;

    #[test]
    fn it_converts_from_string_with_partition() {
        let s = String::from("part:id:id");
        let guid = Guid::from(s.as_str());
        assert_eq!(guid.partition(), Some(String::from("part")));
        assert_eq!(guid.id(), &String::from("id:id"));
    }

    #[test]
    fn it_converts_from_string_without_partition() {
        let s = String::from("no-partition");
        let guid = Guid::from(s.as_str());
        assert_eq!(guid.partition(), None);
        assert_eq!(guid.id(), &s);
    }

    #[test]
    fn it_converts_from_string_and_to_string() {
        let s = String::from("part:id");
        let guid = Guid::from(s.as_str());
        assert_eq!(guid.to_string(), s);
    }

    #[test]
    fn it_serializes() {
        let s = String::from("part:id");
        let guid = Guid::from(s.as_str());
        let json = serde_json::to_string(&guid).unwrap();
        assert_eq!(json, format!("\"{}\"", &s));
    }

    #[test]
    fn it_deserializes() {
        let s = String::from("part:id");
        let guid: Guid = serde_json::from_str("\"part:id\"").unwrap();

        assert_eq!(guid.partition(), Some(String::from("part")));
        assert_eq!(guid.id(), &String::from("id"));
        assert_eq!(guid.to_string(), s);
    }
}