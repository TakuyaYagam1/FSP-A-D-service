const notice = document.querySelector("#notice");
const appShell = document.querySelector("#app-shell");
const sessionState = document.querySelector("#session-state");
const logoutButton = document.querySelector("#logout-button");
const caseForm = document.querySelector("#case-form");
const carrierSelect = document.querySelector("#carrier-select");
const inspectorSelect = document.querySelector("#inspector-select");
const documentSteps = document.querySelectorAll(".document-pin-state span");
const documentDrawer = document.querySelector("#document-drawer");
let currentUsername = "";
let flow = {};

function saveFlow() {
  // Keep only document references; the private code stays on the server.
  try {
    sessionStorage.setItem(
      `wahrwelt.flow.${currentUsername}`,
      JSON.stringify(flow),
    );
  } catch {}
}

function refreshFlow() {
  const hasCase = Boolean(flow.case_id);
  const hasSlot = Boolean(flow.route_slot_id);
  document.querySelector('[data-document-tab="slot"]').disabled = !hasCase;
  document.querySelector('[data-document-tab="approval"]').disabled = !hasSlot;
  document.querySelector("#case-id").value = flow.case_id || "";
  document.querySelector("#slot-id").value = flow.route_slot_id || "";
  document.querySelector("#document-case").textContent = hasCase
    ? "Создано"
    : "Не создано";
  document.querySelector("#document-slot").textContent = hasSlot
    ? "Участники назначены"
    : "Не назначен";
  caseForm.querySelector('button[type="submit"]').disabled = hasCase;
  document.querySelector("#continue-slot").hidden = !hasCase;
  document.querySelector("#new-case-button").hidden = !hasCase;
  caseForm.elements.origin.readOnly = hasCase;
  caseForm.elements.destination.readOnly = hasCase;
  caseForm.elements.kind.disabled = hasCase;
  document.querySelector("#slot-form button[type='submit']").disabled = hasSlot;
  document.querySelector("#flow-progress").textContent =
    `Шаг ${hasSlot ? 3 : hasCase ? 2 : 1} из 3`;
  document.querySelector("#flow-summary").textContent = hasCase
    ? `${flow.origin} -> ${flow.destination}`
    : "Создайте дело, назначьте участников и получите код.";
  if (hasCase) {
    caseForm.elements.kind.value = flow.kind;
    caseForm.elements.origin.value = flow.origin;
    caseForm.elements.destination.value = flow.destination;
    document.querySelector("#document-route").textContent =
      `${flow.origin} -> ${flow.destination}`;
  }
  if (hasSlot) {
    document.querySelector("#completion-summary").textContent =
      `${flow.origin} -> ${flow.destination}. Перевозчик: ${flow.carrier_username}. Инспектор: ${flow.inspector_username}.`;
    document.querySelector("#document-ids").textContent =
      `Дело: ${flow.case_id}. Слот: ${flow.route_slot_id}.`;
    document.querySelector("#approval-help").textContent =
      `Согласование выполняется из аккаунта ${flow.carrier_username} за перевозчика и ${flow.inspector_username} за инспектора. Вы вошли как ${currentUsername}. Выбор роли не меняет ваш аккаунт.`;
  }
  refreshApproval();
  setDocumentStep(hasSlot ? 2 : hasCase ? 1 : 0);
}

function refreshApproval() {
  const role = document.querySelector("#approval-role").value;
  const assigned =
    role === "carrier" ? flow.carrier_username : flow.inspector_username;
  const button = document.querySelector("#approve-button");
  button.disabled = !assigned || assigned !== currentUsername;
  button.textContent =
    button.disabled && assigned
      ? `Нужен аккаунт ${assigned}`
      : "Согласовать свою часть";
}

function showNotice(message, isError = false) {
  notice.textContent = message;
  notice.classList.toggle("error", isError);
}

function formValues(form) {
  return Object.fromEntries(new FormData(form));
}

async function request(path, options = {}) {
  let response;
  try {
    response = await fetch(path, {
      credentials: "same-origin",
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
    });
  } catch {
    throw new Error(
      "Нет связи с сервисом. Проверьте подключение и повторите попытку.",
    );
  }
  if (response.status === 204) return null;
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const messages = {
      bad_request:
        "Проверьте поля формы. Участники должны быть разными пользователями.",
      unauthorized: "Сессия завершена. Войдите в аккаунт заново.",
      forbidden:
        "Ваш аккаунт не назначен на эту роль. Согласовать может только указанный перевозчик или инспектор.",
      not_found: "Документ не найден или недоступен вашему аккаунту.",
      approval_not_ready:
        "Маршрут еще не согласован. Нужны подтверждения назначенных перевозчика и инспектора.",
      conflict:
        "Операция недоступна в текущем состоянии. Проверьте предыдущий шаг.",
    };
    const error = new Error(
      messages[body.code] ||
        "Не удалось выполнить операцию. Повторите попытку позже.",
    );
    error.status = response.status;
    throw error;
  }
  return body;
}

function setSignedIn(username) {
  document.querySelector("#startup-status").hidden = true;
  sessionState.textContent = username;
  appShell.hidden = false;
}

function setUserOptions(users) {
  for (const select of [carrierSelect, inspectorSelect]) {
    const placeholder = select.options[0];
    select.replaceChildren(placeholder);
    for (const user of users) {
      const option = document.createElement("option");
      option.value = user.username;
      option.textContent = user.username;
      select.append(option);
    }
  }
  syncUserOptions();
}

function syncUserOptions(changedSelect) {
  const otherSelect =
    changedSelect === carrierSelect ? inspectorSelect : carrierSelect;
  if (
    changedSelect &&
    changedSelect.value &&
    changedSelect.value === otherSelect.value
  ) {
    otherSelect.value = "";
  }
  for (const option of carrierSelect.options) {
    option.disabled = Boolean(
      option.value && option.value === inspectorSelect.value,
    );
  }
  for (const option of inspectorSelect.options) {
    option.disabled = Boolean(
      option.value && option.value === carrierSelect.value,
    );
  }
}

carrierSelect.addEventListener("change", () => syncUserOptions(carrierSelect));
inspectorSelect.addEventListener("change", () =>
  syncUserOptions(inspectorSelect),
);

function setDocumentStep(index) {
  documentSteps.forEach((step, stepIndex) =>
    step.classList.toggle("active", stepIndex <= index),
  );
}

function showDocumentPane(name) {
  showNotice("");
  document.querySelectorAll("[data-document-tab]").forEach((button) => {
    button.classList.toggle("active", button.dataset.documentTab === name);
  });
  document.querySelectorAll("[data-document-pane]").forEach((pane) => {
    pane.hidden = pane.dataset.documentPane !== name;
  });
}

function openDocuments() {
  appShell.classList.add("documents-open");
  documentDrawer.hidden = false;
  documentDrawer.setAttribute("aria-hidden", "false");
}

function closeDocuments() {
  appShell.classList.remove("documents-open");
  documentDrawer.hidden = true;
  documentDrawer.setAttribute("aria-hidden", "true");
}

document
  .querySelector("#open-documents")
  .addEventListener("click", openDocuments);
document
  .querySelector("#close-documents")
  .addEventListener("click", closeDocuments);
document.querySelectorAll("[data-document-tab]").forEach((button) => {
  button.addEventListener("click", () =>
    showDocumentPane(button.dataset.documentTab),
  );
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") closeDocuments();
});

function selectRoute(routeButton) {
  if (flow.case_id) {
    openDocuments();
    showNotice("Для другого маршрута нажмите 'Начать новую поставку'.");
    return;
  }
  document
    .querySelectorAll("[data-route]")
    .forEach((item) => item.classList.remove("active"));
  document
    .querySelectorAll("[data-route-line]")
    .forEach((item) => item.classList.remove("active"));
  routeButton.classList.add("active");
  const line = document.querySelector(
    `[data-route-line="${routeButton.dataset.route}"]`,
  );
  if (line) line.classList.add("active");

  const route = `Орел -> ${routeButton.dataset.city}`;
  document.querySelector("#document-route").textContent = route;
  caseForm.elements.origin.value = "Орел";
  caseForm.elements.destination.value = routeButton.dataset.city;
}

document.querySelectorAll("button[data-route]").forEach((button) => {
  button.addEventListener("click", () => selectRoute(button));
});

function showResult(target, lines) {
  target.replaceChildren();
  for (const line of lines) {
    const row = document.createElement("p");
    row.textContent = line;
    target.append(row);
  }
  target.hidden = false;
}

async function submitForm(form, action) {
  const buttons = [...form.querySelectorAll("button")];
  if (form.dataset.busy) return;
  form.dataset.busy = "true";
  buttons.forEach((button) => {
    button.disabled = true;
  });
  showNotice("");
  try {
    await action(formValues(form));
  } catch (error) {
    showNotice(error.message, true);
  } finally {
    delete form.dataset.busy;
    buttons.forEach((button) => {
      button.disabled = false;
    });
    refreshFlow();
  }
}

logoutButton.addEventListener("click", async () => {
  logoutButton.disabled = true;
  try {
    await request("/api/v1/logout", { method: "POST" });
    window.location.replace("/login");
  } catch (error) {
    logoutButton.disabled = false;
    showNotice(error.message, true);
  }
});

caseForm.addEventListener("submit", (event) => {
  event.preventDefault();
  void submitForm(event.currentTarget, async (values) => {
    const item = await request("/api/v1/cases", {
      method: "POST",
      body: JSON.stringify(values),
    });
    flow = {
      case_id: item.case_id,
      kind: values.kind,
      origin: item.origin,
      destination: item.destination,
    };
    saveFlow();
    refreshFlow();
    showDocumentPane("slot");
    showNotice("Дело создано. Выберите участников и задайте код поставки.");
  });
});

document.querySelector("#slot-form").addEventListener("submit", (event) => {
  event.preventDefault();
  void submitForm(event.currentTarget, async (values) => {
    const { case_id: caseID, ...body } = values;
    const item = await request(
      `/api/v1/cases/${encodeURIComponent(caseID)}/route-slots`,
      { method: "POST", body: JSON.stringify(body) },
    );
    flow = {
      ...flow,
      route_slot_id: item.route_slot_id,
      carrier_username: item.carrier_username,
      inspector_username: item.inspector_username,
    };
    saveFlow();
    document.querySelector("#slot-form [name='slot_code']").value = "";
    document.querySelector("#approval-role").value =
      currentUsername === item.inspector_username ? "inspector" : "carrier";
    refreshFlow();
    showDocumentPane("approval");
    showNotice("Поставка оформлена. Теперь можно получить код.");
  });
});

document
  .querySelector("#slot-action-form")
  .addEventListener("submit", (event) => {
    event.preventDefault();
    void submitForm(event.currentTarget, async (values) => {
      const result = await request(
        `/api/v1/route-slots/${encodeURIComponent(values.slot_id)}/approvals`,
        {
          method: "POST",
          body: JSON.stringify({ role: values.role }),
        },
      );
      showResult(document.querySelector("#slot-action-result"), [
        `Перевозчик: ${result.carrier_approved ? "согласовано" : "ожидается"}`,
        `Инспектор: ${result.inspector_approved ? "согласовано" : "ожидается"}`,
        `Маршрут готов: ${result.ready ? "да" : "нет"}`,
      ]);
      if (result.ready) setDocumentStep(2);
      showNotice(
        result.ready
          ? "Все согласования получены."
          : "Согласование записано. Требуется подтверждение второй роли.",
      );
    });
  });

document
  .querySelector("#read-code-button")
  .addEventListener("click", async () => {
    const slotID = document.querySelector("#slot-id").value.trim();
    if (!slotID) {
      showNotice("Сначала укажите или создайте маршрутный слот.", true);
      return;
    }
    const button = document.querySelector("#read-code-button");
    button.disabled = true;
    document.querySelector("#code-result").hidden = true;
    showNotice("");
    try {
      const result = await request(
        `/api/v1/route-slots/${encodeURIComponent(slotID)}/code`,
      );
      showResult(document.querySelector("#code-result"), [
        `Код поставки: ${result.slot_code}`,
      ]);
      showNotice("Готово. Код поставки получен.");
    } catch (error) {
      showNotice(error.message, true);
    } finally {
      button.disabled = false;
    }
  });

document
  .querySelector("#approval-role")
  .addEventListener("change", refreshApproval);
document
  .querySelector("#continue-slot")
  .addEventListener("click", () => showDocumentPane("slot"));
document.querySelector("#new-case-button").addEventListener("click", () => {
  flow = {};
  saveFlow();
  caseForm.reset();
  document.querySelector("#slot-form").reset();
  document.querySelectorAll(".result").forEach((result) => {
    result.hidden = true;
  });
  document.querySelectorAll(".approval-details").forEach((details) => {
    details.open = false;
  });
  refreshFlow();
  const selected = document.querySelector("button[data-route].active");
  if (selected) selectRoute(selected);
  showDocumentPane("case");
});

async function restoreSession() {
  let user;
  try {
    user = await request("/api/v1/me");
  } catch (error) {
    if (error.status === 401) {
      window.location.replace("/login");
    } else {
      document.querySelector("#startup-status").hidden = true;
      appShell.hidden = false;
      openDocuments();
      caseForm.querySelector('button[type="submit"]').disabled = true;
      showNotice(
        `${error.message} Обновите страницу, чтобы повторить подключение.`,
        true,
      );
    }
    return;
  }
  currentUsername = user.username;
  setSignedIn(currentUsername);
  try {
    const saved = JSON.parse(
      sessionStorage.getItem(`wahrwelt.flow.${currentUsername}`) || "{}",
    );
    if (saved && typeof saved === "object" && !Array.isArray(saved))
      flow = saved;
  } catch {}
  refreshFlow();
  showDocumentPane(
    flow.route_slot_id ? "approval" : flow.case_id ? "slot" : "case",
  );
  try {
    const directory = await request("/api/v1/users");
    setUserOptions(directory.users);
    carrierSelect.value =
      flow.carrier_username ||
      (directory.users.some((item) => item.username === currentUsername)
        ? currentUsername
        : "carrier");
    inspectorSelect.value =
      flow.inspector_username ||
      (currentUsername === "inspector" ? "carrier" : "inspector");
    syncUserOptions();
  } catch (error) {
    showNotice(`Не удалось загрузить участников. ${error.message}`, true);
  }
}

void restoreSession();
