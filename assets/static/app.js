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
  if (playback.hasAttribute("data-keep-image-visible") && "ResizeObserver" in window) {
    const header = playback.querySelector(".playback-set-header");
    const guidance = playback.querySelector(".playback-guidance");
    const action = document.querySelector("[data-playback-action]");
    const updateLayout = () => {
      // Try the compact layout, but retain at least 10em of space for guidance.
      const viewport = window.visualViewport?.height ?? document.documentElement.clientHeight;
      const style = getComputedStyle(playback);
      const height = viewport - (playback.getBoundingClientRect().top + window.scrollY)
        - parseFloat(style.paddingTop) - action.getBoundingClientRect().height - 16;
      playback.style.setProperty("--set-height", `${Math.max(0, height)}px`);
      playback.classList.add("pinned-image");
      const minimum = parseFloat(getComputedStyle(guidance).fontSize) * 10;
      if (guidance.clientHeight < minimum) playback.classList.remove("pinned-image");
    };
    const observer = new window.ResizeObserver(updateLayout);
    observer.observe(header);
    observer.observe(action);
    window.addEventListener("resize", updateLayout);
    window.visualViewport?.addEventListener("resize", updateLayout);
    updateLayout();
  }

  const copyButton = playback.querySelector("[data-copy-summary]");
  if (copyButton && navigator.clipboard?.writeText) {
    const summary = playback.querySelector("[data-summary-text]");
    const status = playback.querySelector("[data-copy-status]");
    let resetLabel;
    copyButton.hidden = false;
    copyButton.addEventListener("click", async () => {
      clearTimeout(resetLabel);
      copyButton.disabled = true;
      status.textContent = "";
      try {
        await navigator.clipboard.writeText(summary.textContent);
        copyButton.textContent = "Copied";
        status.textContent = "Summary copied.";
        resetLabel = setTimeout(() => { copyButton.textContent = "Copy summary"; }, 2000);
      } catch {
        copyButton.textContent = "Copy summary";
        status.textContent = "Could not copy summary. Select and copy the text manually.";
      } finally {
        copyButton.disabled = false;
      }
    });
  }

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
        if (remaining <= 0) {
          clearInterval(timer);
          if (document.visibilityState === "visible") {
            playback.querySelector("[data-rest-sound]")?.play().catch(() => {});
          }
        }
      }, 1000);
    }
  }
}
