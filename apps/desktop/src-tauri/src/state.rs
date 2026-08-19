use std::process::Child;
use std::sync::Mutex;

#[derive(Clone)]
pub struct LastTunnel {
    pub port: u16,
    pub protocol: String,
}

pub struct TunnelState {
    pub(crate) child: Mutex<Option<Child>>,
    pub last: Mutex<Option<LastTunnel>>,
}

impl Default for TunnelState {
    fn default() -> Self {
        Self {
            child: Mutex::new(None),
            last: Mutex::new(None),
        }
    }
}

impl TunnelState {
    pub(crate) fn stop_child(child_slot: &mut Option<Child>) -> Result<Option<i32>, String> {
        let Some(mut child) = child_slot.take() else {
            return Ok(None);
        };
        if let Some(status) = child.try_wait().map_err(|error| error.to_string())? {
            return Ok(status.code());
        }
        child
            .kill()
            .map_err(|error| format!("stop tunnel CLI: {error}"))?;
        let status = child
            .wait()
            .map_err(|error| format!("wait for tunnel CLI: {error}"))?;
        Ok(status.code())
    }
}
