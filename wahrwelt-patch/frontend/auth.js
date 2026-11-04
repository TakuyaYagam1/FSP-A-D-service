const page = document.body.dataset.authPage;
const form = document.querySelector("#auth-form");
const notice = document.querySelector("#auth-notice");
const submitButton = form.querySelector("button[type='submit']");

function showNotice(message, isError = false) {
  notice.textContent = message;
  notice.classList.toggle("error", isError);
}

function afterLoginPath() {
  const candidate = new URLSearchParams(window.location.search).get("return");
  if (!candidate || !candidate.startsWith("/") || candidate.startsWith("//")) {
    return "/";
  }
  try {
    const target = new URL(candidate, window.location.origin);
    if (target.origin !== window.location.origin) return "/";
    return target.pathname + target.search + target.hash;
  } catch {
    return "/";
  }
}

async function request(path, options = {}) {
  const response = await fetch(path, {
    credentials: "same-origin",
    ...options,
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
  });
  if (response.status === 204) return null;
  const body = await response.json().catch(() => ({}));
  if (!response.ok)
    throw new Error(body.message || "Не удалось выполнить запрос.");
  return body;
}

function preparePage() {
  if (page === "login") {
    const savedUsername = sessionStorage.getItem("wahrwelt.username");
    if (savedUsername) {
      form.elements.username.value = savedUsername;
      sessionStorage.removeItem("wahrwelt.username");
    }
    if (new URLSearchParams(window.location.search).has("created")) {
      showNotice("Учетная запись создана. Войдите, чтобы продолжить.");
    }
  }
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  submitButton.disabled = true;
  showNotice("");

  const values = Object.fromEntries(new FormData(form));
  try {
    if (page === "register") {
      await request("/api/v1/register", {
        method: "POST",
        body: JSON.stringify(values),
      });
      sessionStorage.setItem("wahrwelt.username", values.username);
      window.location.replace("/login?created=1");
      return;
    }

    await request("/api/v1/login", {
      method: "POST",
      body: JSON.stringify(values),
    });
    window.location.replace(afterLoginPath());
  } catch (error) {
    showNotice(error.message, true);
    submitButton.disabled = false;
  }
});

preparePage();
