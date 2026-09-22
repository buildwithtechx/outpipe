use outpipe::protocol::{decode, encode, Envelope, VERSION};
use serde::Deserialize;

#[derive(Deserialize)]
struct Fixtures {
    fixtures: Vec<Fixture>,
}

#[derive(Deserialize)]
struct Fixture {
    name: String,
    valid: bool,
    raw: String,
}

#[test]
fn round_trips_versioned_envelope() {
    let message = Envelope {
        version: VERSION,
        message_type: "heartbeat".into(),
        request_id: Some("r1".into()),
        payload: Some(serde_json::json!({"timestamp": 1})),
    };
    let decoded = decode::<serde_json::Value>(&encode(&message).unwrap()).unwrap();
    assert_eq!(decoded, message);
}

#[test]
fn checks_shared_conformance_fixtures() {
    let fixtures: Fixtures = serde_json::from_str(include_str!(
        "../../../protocol/fixtures/conformance_fixtures.json"
    ))
    .expect("valid conformance fixture file");

    for fixture in fixtures.fixtures {
        let result = decode::<serde_json::Value>(fixture.raw.as_bytes());
        assert_eq!(result.is_ok(), fixture.valid, "fixture {}", fixture.name);
    }
}
