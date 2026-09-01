use outpipe::protocol::{decode, encode, Envelope, VERSION};

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
