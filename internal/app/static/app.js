document.querySelectorAll("[data-auto-submit]").forEach((form) => {
  form.querySelector("select")?.addEventListener("change", () => form.requestSubmit());
  form.querySelector("[data-submit]")?.setAttribute("hidden", "");
});

const playback = document.querySelector("[data-playback]");
if (playback) {
  const slides = [...playback.querySelectorAll("[data-slide]")];
  const completion = playback.querySelector("[data-complete]");
  const action = document.querySelector("[data-playback-action]");
  const nextButton = action.querySelector("[data-playback-next]");
  const choices = new Map();
  let index = 0;
  let timer;
  let completionSent = false;

  const chooseVariant = (exercise) => {
    if (choices.has(exercise)) return choices.get(exercise);
    const preference = playback.dataset.imagePreference;
    let variant = preference;
    if (preference === "random") variant = Math.random() < 0.5 ? "male" : "female";
    if (preference === "alternating") {
      let next = "male";
      try { next = localStorage.getItem("playback-image-next") || "male"; } catch {}
      variant = next === "female" ? "female" : "male";
      try { localStorage.setItem("playback-image-next", variant === "male" ? "female" : "male"); } catch {}
    }
    choices.set(exercise, variant);
    return variant;
  };

  const saveCompletion = async () => {
    const form = playback.querySelector("[data-completion-form]");
    if (!form || completionSent) return;
    completionSent = true;
    const status = completion.querySelector("[data-completion-status]");
    try {
      const response = await fetch(form.action, { method: "POST", body: new FormData(form) });
      if (!response.ok) throw new Error("save failed");
      status.textContent = "Working weights saved.";
    } catch {
      status.textContent = "Working weights were not saved.";
    } finally {
      completion.querySelector("[data-completion-exit]").hidden = false;
    }
  };

  const advance = () => {
    clearInterval(timer);
    index += 1;
    show();
  };

  const show = () => {
    clearInterval(timer);
    slides.forEach((slide) => slide.hidden = true);
    if (index >= slides.length) {
      completion.hidden = false;
      action.hidden = true;
      completion.querySelector("h1").focus();
      saveCompletion();
      window.scrollTo({ top: 0 });
      return;
    }
    const slide = slides[index];
    slide.hidden = false;
    if (slide.dataset.kind === "set") {
      const variant = chooseVariant(slide.dataset.exercise);
      const image = slide.querySelector(".playback-image");
      image.src = image.dataset[`${variant}Src`];
      image.alt = image.dataset[`${variant}Alt`];
      nextButton.textContent = index === slides.length - 1 ? "Finish" : "Next";
    } else {
      nextButton.textContent = "Skip rest";
      const display = slide.querySelector("[data-rest-timer]");
      let remaining = Number(display.dataset.restSeconds);
      if (remaining <= 0) nextButton.textContent = "Next";
      else timer = setInterval(() => {
        remaining -= 1;
        display.textContent = `${remaining} ${remaining === 1 ? "second" : "seconds"}`;
        if (remaining <= 0) {
          clearInterval(timer);
          nextButton.textContent = "Next";
        }
      }, 1000);
    }
    action.hidden = false;
    slide.querySelector("h1").focus();
    window.scrollTo({ top: 0 });
  };

  nextButton.addEventListener("click", advance);
  show();
}
