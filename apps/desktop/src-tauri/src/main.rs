#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod state;

use commands::{tunnel_start, tunnel_status, tunnel_stop, tunnel_version};
use state::TunnelState;
use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .manage(TunnelState::default())
        .invoke_handler(tauri::generate_handler![
            tunnel_start,
            tunnel_stop,
            tunnel_status,
            tunnel_version
        ])
        .build(tauri::generate_context!())
        .expect("error while building Outpipe desktop application")
        .run(|app, event| {
            if let tauri::RunEvent::Exit = event {
                if let Ok(mut child) = app.state::<TunnelState>().child.lock() {
                    if let Err(error) = TunnelState::stop_child(&mut child) {
                        eprintln!("failed to stop tunnel CLI during shutdown: {error}");
                    }
                }
            }
        });
}
