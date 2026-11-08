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
let ownedCases = [];
let ownedCasesHasMore = false;
let ownedCasesBusy = false;
let currentSlotStatus = null;
let currentPassState = null;
let passImageURL = "";
let passPollTimer = 0;
let participantApprovalError = false;
let participantApproval = readParticipantApproval();

function readParticipantApproval() {
  const params = new URLSearchParams(window.location.search);
  const hasApprovalData = ["slot", "role", "assignee"].some((key) => params.has(key));
  if (!hasApprovalData) return null;

  const slotId = params.get("slot")?.trim() || "";
  const role = params.get("role") || "";
  const assignee = params.get("assignee")?.trim() || "";
  if (!slotId || slotId.length > 120 || !["carrier", "inspector"].includes(role) || !assignee || assignee.length > 32) {
    participantApprovalError = true;
    return null;
  }
  return { slotId, role, assignee };
}

function displayRoute(origin, destination) {
  const displayOrigin = origin === "Орел" ? "Орёл" : origin;
  return `${displayOrigin} -> ${destination}`;
}

function currentApprovalRole() {
  if (participantApproval) {
    return participantApproval.assignee === currentUsername
      ? participantApproval.role
      : "";
  }
  if (flow.carrier_username === currentUsername) return "carrier";
  if (flow.inspector_username === currentUsername) return "inspector";
  return "";
}

function saveFlow() {
  try { sessionStorage.setItem(`wahrwelt.flow.${currentUsername}`, JSON.stringify(flow)); } catch {}
}

function participantShareURL(role, username) {
  const url = new URL("/", window.location.origin);
  url.searchParams.set("slot", flow.route_slot_id);
  url.searchParams.set("role", role);
  url.searchParams.set("assignee", username);
  return url.toString();
}

function refreshParticipantLinks() {
  const section = document.querySelector("#participant-links");
  const hasAssignments = Boolean(
    !participantApproval &&
    flow.route_slot_id &&
    flow.carrier_username &&
    flow.inspector_username
  );
  section.hidden = !hasAssignments;
  if (!hasAssignments) return;

  document.querySelector("#carrier-link-label").textContent =
    `Перевозчик: ${flow.carrier_username}`;
  document.querySelector("#carrier-approval-link").value =
    participantShareURL("carrier", flow.carrier_username);
  document.querySelector("#inspector-link-label").textContent =
    `Инспектор: ${flow.inspector_username}`;
  document.querySelector("#inspector-approval-link").value =
    participantShareURL("inspector", flow.inspector_username);
}

function clearParticipantApproval() {
  participantApproval = null;
  participantApprovalError = false;
  const url = new URL(window.location.href);
  for (const key of ["slot", "role", "assignee"]) {
    url.searchParams.delete(key);
  }
  window.history.replaceState(
    null,
    "",
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function loginPathForCurrentPage() {
  const loginURL = new URL("/login", window.location.origin);
  const returnPath =
    window.location.pathname + window.location.search + window.location.hash;
  loginURL.searchParams.set("return", returnPath);
  return loginURL.pathname + loginURL.search;
}

function activeSlotID() {
  return participantApproval?.slotId || flow.route_slot_id || "";
}

function resetRoutePass() {
  currentPassState = null;
  window.clearTimeout(passPollTimer);
  if (passImageURL) URL.revokeObjectURL(passImageURL);
  passImageURL = "";
  document.querySelector("#pass-output").hidden = true;
  const image = document.querySelector("#pass-image");
  image.removeAttribute("src");
  delete image.dataset.passKey;
  document.querySelector("#download-pass").removeAttribute("href");
  document.querySelector("#pass-issued-at").hidden = true;
  document.querySelector("#pass-status").textContent = "Состояние поставки загружается...";
  document.querySelector("#pass-status").hidden = false;
}

function renderRoutePass() {
  const slotID = activeSlotID();
  const view = routePassViewModel(currentPassState, !participantApproval && Boolean(flow.case_id));
  document.querySelector('[data-document-tab="pass"]').disabled = !slotID || !view.enabled;
  document.querySelector("#continue-pass").hidden = !slotID || !view.enabled;
  document.querySelector("#issue-pass").hidden = !view.canIssue;
  document.querySelector("#pass-status").hidden = view.canDownload;
  document.querySelector("#pass-status").textContent = currentPassState
    ? view.status
    : slotID ? "Проверяем состояние поставки..." : "Сначала назначьте участников поставки.";
  const output = document.querySelector("#pass-output");
  const issuedAt = document.querySelector("#pass-issued-at");
  if (!view.canDownload) {
    output.hidden = true;
    issuedAt.hidden = true;
    return;
  }
  const pass = currentPassState.pass;
  const image = document.querySelector("#pass-image");
  const download = document.querySelector("#download-pass");
  try {
    const key = `${pass.route_slot_id}:${pass.issued_at}`;
    if (image.dataset.passKey !== key) {
      const png = buildRoutePassPNG(pass);
      const bytes = Uint8Array.from(atob(png.split(",")[1]), (char) => char.charCodeAt(0));
      const imageURL = URL.createObjectURL(new Blob([bytes], { type: "image/png" }));
      if (passImageURL) URL.revokeObjectURL(passImageURL);
      passImageURL = imageURL;
      image.src = imageURL;
      download.href = imageURL;
      download.download = `route-pass-${pass.route_slot_id}.png`;
      image.dataset.passKey = key;
    }
    issuedAt.textContent = `Выпущен: ${new Date(pass.issued_at).toISOString().replace("T", " ").slice(0, 16)} UTC`;
    issuedAt.hidden = false;
    output.hidden = false;
  } catch (error) {
    issuedAt.hidden = true;
    output.hidden = true;
    document.querySelector("#pass-status").hidden = false;
    document.querySelector("#pass-status").textContent = error.message;
  }
}

async function loadRoutePass() {
  const slotID = activeSlotID();
  if (!slotID) {
    resetRoutePass();
    refreshFlow();
    return;
  }
  const state = await request(`/api/v1/route-slots/${encodeURIComponent(slotID)}/pass`, { cache: "no-store" });
  if (slotID !== activeSlotID()) return;
  currentPassState = state;
  refreshFlow();
  scheduleParticipantPassPoll();
  if (participantApproval && state.issued && state.pass) showDocumentPane("pass");
}

function scheduleParticipantPassPoll() {
  window.clearTimeout(passPollTimer);
  if (!participantApproval || !activeSlotID() || currentPassState?.issued || document.hidden) return;
  passPollTimer = window.setTimeout(() => {
    void loadRoutePass().catch(() => showNotice("Не удалось проверить пропуск. Обновите страницу.", true));
  }, 15000);
}

function refreshFlow() {
  const hasCase = !participantApproval && Boolean(flow.case_id);
  const slotId = participantApproval?.slotId || flow.route_slot_id || "";
  const hasSlot = Boolean(slotId);
  const hasSavedCases = !participantApproval && ownedCases.length > 0;
  document.querySelector('[data-document-tab="slot"]').disabled = !hasCase;
  document.querySelector('[data-document-tab="approval"]').disabled = !hasSlot;
  document.querySelector("#case-id").value = hasCase ? flow.case_id : "";
  document.querySelector("#slot-id").value = slotId;
  document.querySelector("#document-case").textContent = hasCase
    ? "Создано"
    : participantApproval
      ? "По ссылке"
      : hasSavedCases ? "Есть дела" : "Не создано";
  document.querySelector("#document-slot").textContent = hasSlot
    ? participantApproval
      ? "По ссылке"
      : "Назначены"
    : "Не назначены";
  document.querySelector("#approval-heading").textContent = participantApproval
    ? "Подтверждение маршрута"
    : "Участники назначены";
  logoutButton.textContent =
    participantApproval && participantApproval.assignee !== currentUsername
      ? "Сменить аккаунт"
      : "Выйти";
  caseForm.querySelector('button[type="submit"]').disabled = hasCase;
  document.querySelector("#continue-slot").hidden = !hasCase;
  document.querySelector("#new-case-button").hidden = !hasCase;
  document.querySelector("#open-case-list").hidden = Boolean(participantApproval);
  document.querySelector("#owner-approval-panel").hidden = !hasCase || !hasSlot;
  caseForm.elements.origin.readOnly = hasCase;
  caseForm.elements.destination.readOnly = hasCase;
  caseForm.elements.kind.disabled = hasCase;
  document.querySelector("#slot-form button[type='submit']").disabled = hasSlot;
  document.querySelector("#flow-progress").textContent = `Дело: ${hasCase ? "создано" : hasSavedCases ? "выберите сохраненное" : "не создано"} / участники: ${hasSlot ? "назначены" : "не назначены"}`;
  document.querySelector("#flow-summary").textContent = participantApproval
    ? "Подтверждение маршрута по ссылке."
    : hasCase
      ? displayRoute(flow.origin, flow.destination)
      : hasSavedCases
        ? "Выберите дело в «Мои дела» или создайте новое."
        : "Создайте дело, назначьте участников и выпустите пропуск после подтверждений.";
  if (participantApproval) {
    document.querySelector("#document-route").textContent = "Поставка по ссылке";
  }
  if (hasCase) {
    caseForm.elements.kind.value = flow.kind;
    caseForm.elements.origin.value = flow.origin;
    caseForm.elements.destination.value = flow.destination;
    document.querySelector("#document-route").textContent = displayRoute(flow.origin, flow.destination);
  }
  if (hasSlot) {
    if (participantApproval) {
      const roleName = participantApproval.role === "carrier" ? "перевозчик" : "инспектор";
      document.querySelector("#completion-summary").textContent =
        `Поставка ${slotId}. Ссылка назначена на роль "${roleName}" для аккаунта ${participantApproval.assignee}. Сейчас вы вошли как ${currentUsername}.`;
      document.querySelector("#document-ids").textContent = `Поставка: ${slotId}.`;
      document.querySelector("#approval-help").textContent =
        `Ссылка указывает роль "${roleName}" для аккаунта ${participantApproval.assignee}. Ссылка сама по себе не дает доступ. Сервис сверит аккаунт с назначением.`;
    } else {
      document.querySelector("#completion-summary").textContent = `${displayRoute(flow.origin, flow.destination)}. Перевозчик: ${flow.carrier_username}. Инспектор: ${flow.inspector_username}.`;
      document.querySelector("#document-ids").textContent = `Дело: ${flow.case_id}. Поставка: ${flow.route_slot_id}.`;
      document.querySelector("#approval-help").textContent = `Перевозчик: ${flow.carrier_username}. Инспектор: ${flow.inspector_username}. Текущий аккаунт: ${currentUsername}. Подтвердить может только назначенный участник. Сервис сверит аккаунт с назначением.`;
    }
  }
  refreshParticipantLinks();
  refreshApproval();
  renderOwnerStatus();
  renderRoutePass();
  setDocumentStep(currentPassState?.issued ? 3 : hasSlot ? 2 : hasCase ? 1 : 0);
}

function refreshApproval() {
  const role = currentApprovalRole();
  document.querySelector("#approval-role").value = role;
  const button = document.querySelector("#approve-button");
  button.disabled = !role;
  button.textContent = role
    ? `Подтвердить как ${role === "carrier" ? "перевозчик" : "инспектор"}`
    : "Войдите как назначенный участник";
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
      headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    });
  } catch {
    throw new Error("Нет связи с сервисом. Проверьте подключение и повторите попытку.");
  }
  if (response.status === 204) return null;
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const messages = {
      bad_request: "Проверьте поля формы. Участники должны быть разными пользователями.",
      unauthorized: "Сессия завершена. Войдите в аккаунт заново.",
      forbidden: "У текущего аккаунта нет права на это действие.",
      not_found: "Дело или поставка недоступны текущему аккаунту.",
      approval_not_ready: "Маршрут пока не готов к этому действию. Обновите подтверждения поставки.",
      conflict: "Операция недоступна в текущем состоянии. Проверьте предыдущий шаг.",
    };
    const error = new Error(messages[body.code] || "Не удалось выполнить операцию. Повторите попытку позже.");
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
  const otherSelect = changedSelect === carrierSelect ? inspectorSelect : carrierSelect;
  if (changedSelect && changedSelect.value && changedSelect.value === otherSelect.value) {
    otherSelect.value = "";
  }
  for (const option of carrierSelect.options) {
    option.disabled = Boolean(option.value && option.value === inspectorSelect.value);
  }
  for (const option of inspectorSelect.options) {
    option.disabled = Boolean(option.value && option.value === carrierSelect.value);
  }
}

carrierSelect.addEventListener("change", () => syncUserOptions(carrierSelect));
inspectorSelect.addEventListener("change", () => syncUserOptions(inspectorSelect));

function setDocumentStep(index) {
  documentSteps.forEach((step, stepIndex) => step.classList.toggle("active", stepIndex <= index));
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

document.querySelector("#open-documents").addEventListener("click", openDocuments);
document.querySelector("#close-documents").addEventListener("click", closeDocuments);
document.querySelectorAll("[data-document-tab]").forEach((button) => {
  button.addEventListener("click", () => {
    showDocumentPane(button.dataset.documentTab);
    if (button.dataset.documentTab === "pass") {
      void loadRoutePass().catch((error) => showNotice(error.message, true));
    }
  });
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") closeDocuments();
});
document.addEventListener("visibilitychange", () => {
  if (document.hidden) {
    window.clearTimeout(passPollTimer);
  } else if (participantApproval && activeSlotID() && !currentPassState?.issued) {
    void loadRoutePass().catch((error) => showNotice(error.message, true));
  }
});

function selectRoute(routeButton) {
  if (flow.case_id) {
    openDocuments();
    showNotice("Для другого маршрута создайте новое дело.");
    return;
  }
  document.querySelectorAll("[data-route]").forEach((item) => item.classList.remove("active"));
  document.querySelectorAll("[data-route-line]").forEach((item) => item.classList.remove("active"));
  routeButton.classList.add("active");
  document
    .querySelectorAll(`[data-route-line="${routeButton.dataset.route}"]`)
    .forEach((line) => line.classList.add("active"));

  const route = displayRoute("Орел", routeButton.dataset.city);
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
  buttons.forEach((button) => { button.disabled = true; });
  showNotice("");
  try {
    await action(formValues(form));
  } catch (error) {
    showNotice(error.message, true);
  } finally {
    delete form.dataset.busy;
    buttons.forEach((button) => { button.disabled = false; });
    refreshFlow();
  }
}

logoutButton.addEventListener("click", async () => {
  logoutButton.disabled = true;
  try {
    await request("/api/v1/logout", { method: "POST" });
    window.location.replace(loginPathForCurrentPage());
  } catch (error) {
    logoutButton.disabled = false;
    showNotice(error.message, true);
  }
});

caseForm.addEventListener("submit", (event) => {
  event.preventDefault();
  void submitForm(event.currentTarget, async (values) => {
    const item = await request("/api/v1/cases", { method: "POST", body: JSON.stringify(values) });
    clearParticipantApproval();
    resetRoutePass();
    flow = { case_id: item.case_id, kind: values.kind, origin: item.origin, destination: item.destination };
    currentSlotStatus = null;
    saveFlow();
    void loadOwnedCases(true).catch((error) => showNotice(`Дело создано, но список не обновился. ${error.message}`, true));
    refreshFlow();
    showDocumentPane("slot");
    showNotice("Дело создано. Назначьте перевозчика и инспектора, затем задайте код поставки.");
  });
});

document.querySelector("#slot-form").addEventListener("submit", (event) => {
  event.preventDefault();
  void submitForm(event.currentTarget, async (values) => {
    const { case_id: caseID, ...body } = values;
    const item = await request(`/api/v1/cases/${encodeURIComponent(caseID)}/route-slots`, { method: "POST", body: JSON.stringify(body) });
    resetRoutePass();
    flow = { ...flow, route_slot_id: item.route_slot_id, carrier_username: item.carrier_username, inspector_username: item.inspector_username };
    currentSlotStatus = { slot: item, approval_count: 0, ready: false };
    saveFlow();
    void loadOwnedCases(true).catch((error) => showNotice(`Назначения сохранены, но список не обновился. ${error.message}`, true));
    document.querySelector("#slot-form [name='slot_code']").value = "";
    refreshFlow();
    showDocumentPane("approval");
    showNotice("Участники назначены. Код поставки можно получить.");
  });
});

document.querySelector("#slot-action-form").addEventListener("submit", (event) => {
  event.preventDefault();
  void submitForm(event.currentTarget, async (values) => {
    const role = currentApprovalRole();
    if (!role) throw new Error("Ваш аккаунт не назначен участником этой поставки.");
    document.querySelector("#approval-role").value = role;
    const result = await request(`/api/v1/route-slots/${encodeURIComponent(values.slot_id)}/approvals`, {
      method: "POST",
      body: JSON.stringify({ role }),
    });
    showResult(document.querySelector("#slot-action-result"), [
      "Подтверждение записано.",
      `Маршрут готов: ${result.ready ? "да" : "нет"}`,
    ]);
    showNotice(result.ready ? "Маршрут отмечен как готовый. Владелец может выпустить пропуск." : "Подтверждение записано. Маршрут пока не готов.");
    void loadRoutePass().catch((error) => showNotice(error.message, true));
  });
});

document.querySelectorAll("[data-copy-link]").forEach((button) => {
  button.addEventListener("click", async () => {
    const input = document.getElementById(button.dataset.copyLink);
    const linkNotice = document.querySelector("#link-notice");
    try {
      if (!navigator.clipboard?.writeText) throw new Error("Clipboard unavailable");
      await navigator.clipboard.writeText(input.value);
      linkNotice.textContent = "Ссылка скопирована.";
    } catch {
      input.focus();
      input.select();
      linkNotice.textContent = "Скопируйте выделенный URL вручную.";
    }
  });
});

document.querySelector("#read-code-button").addEventListener("click", async () => {
  const slotID = document.querySelector("#slot-id").value.trim();
  if (!slotID) {
    showNotice("Сначала создайте дело и назначьте участников.", true);
    return;
  }
  const button = document.querySelector("#read-code-button");
  button.disabled = true;
  document.querySelector("#code-result").hidden = true;
  showNotice("");
  try {
    const result = await request(`/api/v1/route-slots/${encodeURIComponent(slotID)}/code`);
    showResult(document.querySelector("#code-result"), [`Код поставки: ${result.slot_code}`]);
    showNotice("Готово. Код поставки получен.");
  } catch (error) {
    showNotice(error.message, true);
  } finally {
    button.disabled = false;
  }
});

document.querySelector("#continue-slot").addEventListener("click", () => showDocumentPane("slot"));

document.querySelector("#continue-pass").addEventListener("click", () => {
  showDocumentPane("pass");
  void loadRoutePass().catch((error) => showNotice(error.message, true));
});
document.querySelector("#refresh-pass").addEventListener("click", async () => {
  const button = document.querySelector("#refresh-pass");
  button.disabled = true;
  try {
    await loadRoutePass();
  } catch (error) {
    showNotice(error.message, true);
  } finally {
    button.disabled = false;
  }
});
document.querySelector("#issue-pass").addEventListener("click", async () => {
  const button = document.querySelector("#issue-pass");
  const slotID = activeSlotID();
  if (!slotID || participantApproval || !flow.case_id) return;
  button.disabled = true;
  try {
    const state = await request(`/api/v1/route-slots/${encodeURIComponent(slotID)}/pass`, { method: "POST" });
    if (slotID !== activeSlotID()) return;
    currentPassState = state;
    refreshFlow();
    showDocumentPane("pass");
    showNotice("Пропуск выпущен. Назначенные участники могут скачать его по своим ссылкам.");
  } catch (error) {
    showNotice(error.message, true);
  } finally {
    button.disabled = false;
  }
});

document.querySelector("#new-case-button").addEventListener("click", () => {
  resetRoutePass();
  flow = {};
  currentSlotStatus = null;
  saveFlow();
  caseForm.reset();
  document.querySelector("#slot-form").reset();
  document.querySelectorAll(".result").forEach((result) => { result.hidden = true; });
  document.querySelectorAll(".approval-details").forEach((details) => { details.open = false; });
  refreshFlow();
  renderOwnedCases();
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
      window.location.replace(loginPathForCurrentPage());
    } else {
      document.querySelector("#startup-status").hidden = true;
      appShell.hidden = false;
      openDocuments();
      caseForm.querySelector('button[type="submit"]').disabled = true;
      showNotice(`${error.message} Обновите страницу, чтобы повторить подключение.`, true);
    }
    return;
  }
  currentUsername = user.username;
  setSignedIn(currentUsername);
  try {
    const saved = JSON.parse(sessionStorage.getItem(`wahrwelt.flow.${currentUsername}`) || "{}");
    if (saved && typeof saved === "object" && !Array.isArray(saved)) flow = saved;
  } catch {}
  refreshFlow();
  showDocumentPane(
    participantApproval || flow.route_slot_id
      ? "approval"
      : flow.case_id
        ? "slot"
        : "case",
  );
  if (participantApproval) {
    const approvalDetails = document.querySelector(".approval-details");
    approvalDetails.open = true;
    openDocuments();
    void loadRoutePass().catch((error) => showNotice(error.message, true));
  }
  if (participantApprovalError) {
    showNotice("Ссылка для подтверждения неполная или некорректная.", true);
  }
  try {
    const directory = await request("/api/v1/users");
    setUserOptions(directory.users);
    carrierSelect.value = flow.carrier_username || "";
    inspectorSelect.value = flow.inspector_username || "";
    syncUserOptions();
  } catch (error) {
    showNotice(`Не удалось загрузить участников. ${error.message}`, true);
  }
  if (!participantApproval) {
    const savedCaseID = flow.case_id;
    const savedSlotID = flow.route_slot_id;
    try {
      await loadOwnedCases(true);
      if (savedCaseID) {
        let entry = ownedCases.find((item) => item.case.case_id === savedCaseID);
        if (!entry) entry = { case: await request(`/api/v1/cases/${encodeURIComponent(savedCaseID)}`) };
        await openOwnedCase(entry, savedSlotID, false);
      } else if (ownedCases.length) {
        showDocumentPane("cases");
      }
    } catch (error) {
      if (error.status === 404 && savedCaseID) {
        flow = {};
        currentSlotStatus = null;
        saveFlow();
        refreshFlow();
        renderOwnedCases();
        showDocumentPane("cases");
      }
      showNotice(`Не удалось восстановить дела. ${error.message}`, true);
    }
  }
}

void restoreSession();
