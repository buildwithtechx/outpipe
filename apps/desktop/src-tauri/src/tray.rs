use tauri::menu::{Menu, MenuItem};
use tauri::tray::TrayIconBuilder;
use tauri::{App, AppHandle, Manager};

use crate::commands::tunnel::{spawn_tunnel, stop_child_for, TunnelOptions};
use crate::notify::notify;
use crate::state::TunnelState;

pub const TRAY_ID: &str = "outpipe-tray";

pub fn create_tray(app: &App) -> Result<(), Box<dyn std::error::Error>> {
    let Some(icon) = app.default_window_icon().cloned() else {
        return Err("application icon is unavailable for the system tray".into());
    };

    TrayIconBuilder::with_id(TRAY_ID)
        .icon(icon)
        .tooltip("Outpipe Tunnel")
        .menu(&menu(app.handle(), "Tunnel: stopped")?)
        .show_menu_on_left_click(true)
        .on_menu_event(|app, event| match event.id().as_ref() {
            "start" => start_from_tray(app),
            "stop" => stop_from_tray(app),
            "quit" => app.exit(0),
            _ => {}
        })
        .build(app)?;

    Ok(())
}

fn menu(app: &AppHandle, status: &str) -> tauri::Result<Menu<tauri::Wry>> {
    Menu::with_items(
        app,
        &[
            &MenuItem::with_id(app, "status", status, false, None::<&str>)?,
            &MenuItem::with_id(app, "start", "Start Last Tunnel", true, None::<&str>)?,
            &MenuItem::with_id(app, "stop", "Stop Tunnel", true, None::<&str>)?,
            &MenuItem::with_id(app, "quit", "Quit Outpipe", true, None::<&str>)?,
        ],
    )
}

pub fn set_status(app: &AppHandle, status: &str) {
    let Ok(menu) = menu(app, status) else {
        return;
    };
    let Some(tray) = app.tray_by_id(TRAY_ID) else {
        return;
    };
    tray.set_menu(Some(menu))
        .map_err(|error| eprintln!("update tray menu: {error}"))
        .ok();
}

fn start_from_tray(app: &AppHandle) {
    let Some(state) = app
        .state::<TunnelState>()
        .last
        .lock()
        .ok()
        .and_then(|guard| guard.clone())
    else {
        let _ = notify(
            app,
            "No tunnel configuration saved",
            "Start a tunnel from the Outpipe dashboard once, then the tray can restart it.",
        );
        return;
    };
    match spawn_tunnel(app, state.port, state.protocol, TunnelOptions::default()) {
        Ok(process) => set_status(app, &format!("Tunnel: {}", process.status)),
        Err(error) => {
            let _ = notify(app, "Tunnel failed to start", &error);
            set_status(app, "Tunnel: stopped");
        }
    }
}

fn stop_from_tray(app: &AppHandle) {
    match stop_child_for(app) {
        Ok(_) => {
            let _ = notify(app, "Tunnel stopped", "The CLI tunnel client was stopped.");
            set_status(app, "Tunnel: stopped");
        }
        Err(error) => {
            let _ = notify(app, "Failed to stop tunnel", &error);
        }
    }
}
