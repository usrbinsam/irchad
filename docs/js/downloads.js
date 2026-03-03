document$.subscribe(function () {
  const btn = document.querySelector(".download-btn");
  if (!btn) return;

  const platform = navigator.userAgent.toLowerCase();
  const baseUrl =
    "https://github.com/usrbinsam/irchad/releases/latest/download/";

  if (platform.includes("win")) {
    btn.href = baseUrl + "irchad-windows.tar.xz";
    btn.innerText = "Download for Windows";
  } else if (platform.includes("linux")) {
    btn.href = baseUrl + "irchad-linux.tar.xz";
    btn.innerText = "Download for Linux";
  } else {
    btn.href = "https://github.com/usrbinsam/irchad/releases/latest";
    btn.innerText = "View All Downloads";
  }
});
