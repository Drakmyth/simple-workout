if ("serviceWorker" in navigator) {
  window.addEventListener("load", () => {
    navigator.serviceWorker.register("/service-worker.js").catch(() => {});
  });
}

document.querySelectorAll("[data-auto-submit]").forEach((form) => {
  form.querySelector("select")?.addEventListener("change", () => form.requestSubmit());
  form.querySelector("[data-submit]")?.setAttribute("hidden", "");
});

const playback = document.querySelector("[data-playback]");
if (playback) {
  if (playback.hasAttribute("data-keep-screen-on") && "wakeLock" in navigator) {
    let lock = null;
    let requesting = false;
    let active = true;

    const requestLock = async () => {
      if (!active || document.visibilityState !== "visible" || lock || requesting) return;
      requesting = true;
      try {
        const acquired = await navigator.wakeLock.request("screen");
        if (!active || document.visibilityState !== "visible") {
          await acquired.release();
        } else {
          lock = acquired;
          acquired.addEventListener("release", () => {
            if (lock === acquired) lock = null;
          });
        }
      } catch {
        // Wake lock is best-effort; unsupported or denied requests never block playback.
      } finally {
        requesting = false;
      }
    };

    document.addEventListener("visibilitychange", requestLock);
    window.addEventListener("pageshow", () => {
      active = true;
      requestLock();
    });
    window.addEventListener("pagehide", () => {
      active = false;
      const acquired = lock;
      lock = null;
      if (acquired) acquired.release().catch(() => {});
    });
    requestLock();
  }

  const display = playback.querySelector("[data-rest-timer]");
  if (display) {
    let remaining = Number(display.dataset.restSeconds);
    if (remaining > 0) {
      const timer = setInterval(() => {
        remaining -= 1;
        display.textContent = `${remaining} ${remaining === 1 ? "second" : "seconds"}`;
        if (remaining <= 0) clearInterval(timer);
      }, 1000);
    }
  }
}
