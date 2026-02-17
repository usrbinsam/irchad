package hack

/*
#cgo linux pkg-config: gtk+-3.0 webkit2gtk-4.1
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>

static gboolean on_permission_request(WebKitWebView *v, WebKitPermissionRequest *r, gpointer d) {
    webkit_permission_request_allow(r);
    return TRUE;
}

static void patch_webview(GtkWidget *widget, gpointer data) {
    if (WEBKIT_IS_WEB_VIEW(widget)) {
        // Force hardware flags on
        WebKitSettings *settings = webkit_web_view_get_settings(WEBKIT_WEB_VIEW(widget));
        g_object_set(settings, "enable-media-stream", TRUE, "enable-webrtc", TRUE, NULL);

        // Intercept the prompt
        g_signal_connect(widget, "permission-request", G_CALLBACK(on_permission_request), NULL);
    } else if (GTK_IS_CONTAINER(widget)) {
        gtk_container_forall(GTK_CONTAINER(widget), (GtkCallback)patch_webview, NULL);
    }
}

void HackAllowGetUserMedia(void *handle) {
    if (handle != 0) {
        patch_webview(GTK_WIDGET((void*)handle), NULL);
    }
}
*/
import "C"
import "unsafe"

// HackAllowGetUserMedia takes a pointer to the main application window
// and enables access to navigator.mediaDevices.getUserMedia()
//
// This is only needed until Wails v3 exposes this functionality natively.
func HackAllowGetUserMedia(ptr unsafe.Pointer) {
	C.HackAllowGetUserMedia(ptr)
}
