use tauri::AppHandle;
use tauri_plugin_notification::NotificationExt;

pub fn notify(app: &AppHandle, title: &str, body: &str) -> Result<(), String> {
    let notification = app
        .notification()
        .builder()
        .title(title.to_string())
        .body(body.to_string());
    notification
        .show()
        .map_err(|error| format!("send notification: {error}"))
}
