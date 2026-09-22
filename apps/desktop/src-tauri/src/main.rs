#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod credentials;
mod notify;
mod state;
mod tray;

use commands::{tunnel_start, tunnel_status, tunnel_stop, tunnel_version};
use credentials::{credentials_clear, credentials_save, credentials_status};
use state::TunnelState;
use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .manage(TunnelState::default())
        .plugin(tauri_plugin_notification::init())
        .invoke_handler(tauri::generate_handler![
            tunnel_start,
            tunnel_stop,
            tunnel_status,
            tunnel_version,
            credentials_save,
            credentials_clear,
            credentials_status
        ])
        .setup(|app| {
            tray::create_tray(app)?;
            Ok(())
        })
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
