use serde::Serialize;
use std::io::Read;
use std::path::PathBuf;
use std::process::{Command, Stdio};
use tauri::{AppHandle, Manager, State};

use crate::credentials::{credential, CredentialKind};
use crate::notify::notify;
use crate::state::{LastTunnel, TunnelState};
use crate::tray::set_status;

#[derive(Serialize)]
pub struct TunnelProcess {
    pub pid: u32,
    pub status: String,
    pub exit_code: Option<i32>,
}

#[derive(Clone, Default)]
pub struct TunnelOptions {
    pub subdomain: Option<String>,
    pub password: Option<String>,
}

#[tauri::command]
pub fn tunnel_start(
    app: AppHandle,
    state: State<'_, TunnelState>,
    port: u16,
    protocol: String,
    subdomain: Option<String>,
    password: Option<String>,
) -> Result<TunnelProcess, String> {
    validate_protocol(&protocol)?;
    let password = validate_password(password)?;
    let options = TunnelOptions {
        subdomain,
        password,
    };
    let tunnel = spawn_tunnel(&app, port, protocol.clone(), options)?;
    if let Ok(mut last) = state.last.lock() {
        *last = Some(LastTunnel { port, protocol });
    }
    set_status(&app, &format!("Tunnel: {}", tunnel.status));
    Ok(tunnel)
}

pub fn spawn_tunnel(
    app: &AppHandle,
    port: u16,
    protocol: String,
    options: TunnelOptions,
) -> Result<TunnelProcess, String> {
    validate_protocol(&protocol)?;
    let state = app.state::<TunnelState>();
    let mut child_slot = state
        .child
        .lock()
        .map_err(|_| "tunnel state is unavailable")?;
    if let Some(child) = child_slot.as_mut() {
        if child
            .try_wait()
            .map_err(|error| error.to_string())?
            .is_none()
        {
            return Err("a tunnel process is already running".to_string());
        }
    }
    *child_slot = None;

    let mut command = tunnel_command(app);
    command
        .arg("open")
        .arg("--port")
        .arg(port.to_string())
        .arg("--protocol")
        .arg(protocol)
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(subdomain) = options.subdomain.filter(|value| !value.trim().is_empty()) {
        command.arg("--subdomain").arg(subdomain);
    }
    if let Some(password) = options.password {
        command.arg("--password").arg(password);
    }
    if let Some(api_key) = credential(CredentialKind::ApiKey) {
        command.env("OUTPIPE_API_KEY", api_key);
    }
    if let Some(agent_token) = credential(CredentialKind::AgentToken) {
        command.env("OUTPIPE_AGENT_TOKEN", agent_token);
    }

    let mut child = command
        .spawn()
        .map_err(|error| format!("start tunnel CLI: {error}"))?;
    let pid = child.id();
    let stderr = child.stderr.take();
    *child_slot = Some(child);

    let monitor = app.clone();
    std::thread::spawn(move || {
        let mut output = String::new();
        if let Some(mut stderr) = stderr {
            if stderr.read_to_string(&mut output).is_err() {
                output.clear();
            }
        }
        let tail = explain_tail(&output);
        let message = if tail.is_empty() {
            "The tunnel client process ended.".to_string()
        } else {
            format!("Tunnel stopped.{tail}")
        };
        if let Err(error) = notify(&monitor, "Outpipe tunnel ended", &message) {
            eprintln!("notification failed: {error}");
        }
        set_status(&monitor, "Tunnel: stopped");
    });

    Ok(TunnelProcess {
        pid,
        status: "running".to_string(),
        exit_code: None,
    })
}

fn explain_tail(stderr: &str) -> String {
    let lines: Vec<&str> = stderr
        .lines()
        .filter(|line| !line.trim().is_empty())
        .collect();
    let tail: String = lines
        .iter()
        .rev()
        .take(3)
        .collect::<Vec<_>>()
        .into_iter()
        .rev()
        .cloned()
        .collect::<Vec<_>>()
        .join(" ");
    if tail.trim().is_empty() {
        String::new()
    } else {
        format!(" {tail}")
    }
}

pub fn stop_child_for(app: &AppHandle) -> Result<Option<i32>, String> {
    let state = app.state::<TunnelState>();
    let mut child_slot = state
        .child
        .lock()
        .map_err(|_| "tunnel state is unavailable")?;
    TunnelState::stop_child(&mut child_slot)
}

#[tauri::command]
pub fn tunnel_stop(app: AppHandle) -> Result<TunnelProcess, String> {
    let exit_code = stop_child_for(&app)?;
    set_status(&app, "Tunnel: stopped");
    Ok(stopped_process(exit_code))
}

#[tauri::command]
pub fn tunnel_status(state: State<'_, TunnelState>) -> Result<TunnelProcess, String> {
    let mut child_slot = state
        .child
        .lock()
        .map_err(|_| "tunnel state is unavailable")?;
    let Some(child) = child_slot.as_mut() else {
        return Ok(stopped_process(None));
    };
    match child.try_wait().map_err(|error| error.to_string())? {
        Some(status) => {
            let result = stopped_process(status.code());
            *child_slot = None;
            Ok(result)
        }
        None => Ok(TunnelProcess {
            pid: child.id(),
            status: "running".to_string(),
            exit_code: None,
        }),
    }
}

#[tauri::command]
pub fn tunnel_version(app: AppHandle) -> Result<String, String> {
    let output = tunnel_command(&app)
        .arg("version")
        .output()
        .map_err(|error| format!("run tunnel CLI: {error}"))?;
    if !output.status.success() {
        return Err(String::from_utf8_lossy(&output.stderr).trim().to_string());
    }
    Ok(String::from_utf8_lossy(&output.stdout).trim().to_string())
}

fn tunnel_command(app: &AppHandle) -> Command {
    Command::new(cli_path(app))
}

fn cli_path(app: &AppHandle) -> PathBuf {
    if let Ok(configured) = std::env::var("OUTPIPE_CLI_PATH") {
        return PathBuf::from(configured);
    }
    if let Ok(resource) = app
        .path()
        .resolve("outpipe", tauri::path::BaseDirectory::Resource)
    {
        if resource.exists() {
            return resource;
        }
    }
    if let Ok(exe) = std::env::current_exe() {
        if let Some(directory) = exe.parent() {
            let adjacent = directory.join(if cfg!(windows) {
                "outpipe.exe"
            } else {
                "outpipe"
            });
            if adjacent.exists() {
                return adjacent;
            }
        }
    }
    PathBuf::from("outpipe")
}

fn validate_password(password: Option<String>) -> Result<Option<String>, String> {
    let Some(value) = password else {
        return Ok(None);
    };

    if value.trim().is_empty() {
        return Err("tunnel password cannot contain only whitespace".to_string());
    }

    let length = value.len();

    if !(8..=256).contains(&length) {
        return Err("tunnel password must be between 8 and 256 characters".to_string());
    }

    Ok(Some(value))
}

fn validate_protocol(protocol: &str) -> Result<(), String> {
    match protocol {
        "http" | "https" | "tcp" | "udp" => Ok(()),
        _ => Err(format!("unsupported tunnel protocol: {protocol}")),
    }
}

fn stopped_process(exit_code: Option<i32>) -> TunnelProcess {
    TunnelProcess {
        pid: 0,
        status: "stopped".to_string(),
        exit_code,
    }
}
