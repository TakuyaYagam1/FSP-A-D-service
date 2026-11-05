"use strict";

function approvalStatusText(status) {
  return `Подтверждений записано: ${status.approval_count}. ${status.ready ? "Маршрут готов." : "Маршрут ожидает подтверждений."}`;
}

function renderOwnerStatus() {
  const output = document.querySelector("#owner-approval-status");
  if (!flow.route_slot_id || participantApproval) return;
  output.textContent = currentSlotStatus?.slot.route_slot_id === flow.route_slot_id
    ? approvalStatusText(currentSlotStatus)
    : "Обновите подтверждения, чтобы узнать текущее состояние.";
}

function caseKindLabel(kind) {
  return { import: "Ввоз", export: "Вывоз", service: "Сервисный рейс" }[kind] || "Дело";
}

function caseListError(error) {
  const message = `Не удалось открыть дело. ${error.message}`;
  document.querySelector("#case-list-status").textContent = message;
  showNotice(message, true);
}

function renderOwnedCases() {
  const list = document.querySelector("#owned-cases");
  list.replaceChildren();
  for (const entry of ownedCases) {
    const item = document.createElement("li");
    item.className = "owned-case";

    const title = document.createElement("h4");
    title.textContent = `${caseKindLabel(entry.case.kind)}: ${displayRoute(entry.case.origin, entry.case.destination)}`;
    item.append(title);

    const id = document.createElement("p");
    id.className = "owned-case-id";
    id.textContent = `Дело ${entry.case.case_id}`;
    item.append(id);

    if (entry.route_slots.length) {
      const slots = document.createElement("div");
      slots.className = "owned-case-slots";
      for (const status of entry.route_slots) {
        const row = document.createElement("div");
        row.className = "owned-case-slot";
        const people = document.createElement("p");
        people.textContent = `Перевозчик: ${status.slot.carrier_username}. Инспектор: ${status.slot.inspector_username}.`;
        const approvals = document.createElement("p");
        approvals.textContent = approvalStatusText(status);
        const button = document.createElement("button");
        button.type = "button";
        button.className = "button button-secondary";
        button.textContent = "Открыть поставку";
        button.addEventListener("click", () => {
          void openOwnedCase(entry, status.slot.route_slot_id).catch(caseListError);
        });
        row.append(people, approvals, button);
        slots.append(row);
      }
      item.append(slots);
    } else {
      const note = document.createElement("p");
      note.textContent = "Участники пока не назначены.";
      item.append(note);
      const button = document.createElement("button");
      button.type = "button";
      button.className = "button button-secondary";
      button.textContent = "Открыть дело";
      button.addEventListener("click", () => {
        void openOwnedCase(entry).catch(caseListError);
      });
      item.append(button);
    }
    list.append(item);
  }
  document.querySelector("#load-more-cases").hidden = !ownedCasesHasMore;
  if (!flow.case_id && !participantApproval) {
    document.querySelector("#open-documents").firstChild.textContent = ownedCases.length ? "Мои дела " : "Открыть дело ";
  }
}

async function loadOwnedCases(reset = true) {
  if (ownedCasesBusy) return;
  ownedCasesBusy = true;
  const status = document.querySelector("#case-list-status");
  const refresh = document.querySelector("#refresh-case-list");
  const more = document.querySelector("#load-more-cases");
  refresh.disabled = true;
  more.disabled = true;
  status.textContent = "Загрузка дел...";
  try {
    const offset = reset ? 0 : ownedCases.length;
    const result = await request(`/api/v1/cases?offset=${offset}`);
    ownedCases = reset ? result.cases : [...ownedCases, ...result.cases];
    ownedCasesHasMore = result.has_more;
    renderOwnedCases();
    if (!flow.case_id && !participantApproval) refreshFlow();
    status.textContent = ownedCases.length ? "" : "Дел пока нет. Создайте первое дело во вкладке «Дело».";
  } catch (error) {
    status.textContent = `Не удалось загрузить дела. ${error.message}`;
    throw error;
  } finally {
    ownedCasesBusy = false;
    refresh.disabled = false;
    more.disabled = false;
  }
}

async function openOwnedCase(entry, selectedSlotID = "", reveal = true) {
  const caseItem = entry.case;
  const result = await request(`/api/v1/cases/${encodeURIComponent(caseItem.case_id)}/route-slots`);
  const statuses = result.route_slots;
  const selected = statuses.find((status) => status.slot.route_slot_id === selectedSlotID) || statuses[0] || null;
  resetRoutePass();
  flow = {
    case_id: caseItem.case_id,
    kind: caseItem.kind,
    origin: caseItem.origin,
    destination: caseItem.destination,
  };
  if (selected) {
    flow.route_slot_id = selected.slot.route_slot_id;
    flow.carrier_username = selected.slot.carrier_username;
    flow.inspector_username = selected.slot.inspector_username;
  }
  currentSlotStatus = selected;
  saveFlow();
  document.querySelectorAll(".result").forEach((node) => { node.hidden = true; });
  refreshFlow();
  document.querySelector("#open-documents").firstChild.textContent = "Открыть дело ";
  if (reveal) {
    showDocumentPane(selected ? "approval" : "slot");
    openDocuments();
  }
  if (selected) void loadRoutePass().catch((error) => showNotice(error.message, true));
}

async function refreshOwnedCaseStatus() {
  if (!flow.case_id || !flow.route_slot_id || participantApproval) return;
  const result = await request(`/api/v1/cases/${encodeURIComponent(flow.case_id)}/route-slots`);
  currentSlotStatus = result.route_slots.find((status) => status.slot.route_slot_id === flow.route_slot_id) || null;
  if (!currentSlotStatus) throw new Error("Поставка не найдена в этом деле.");
  const entry = ownedCases.find((item) => item.case.case_id === flow.case_id);
  if (entry) {
    entry.route_slots = result.route_slots;
    renderOwnedCases();
  }
  renderOwnerStatus();
  await loadRoutePass();
}

document.querySelector("#open-case-list").addEventListener("click", () => {
  showDocumentPane("cases");
  void loadOwnedCases(true).catch((error) => showNotice(error.message, true));
});
document.querySelector("#refresh-case-list").addEventListener("click", () => {
  void loadOwnedCases(true).catch((error) => showNotice(error.message, true));
});
document.querySelector("#load-more-cases").addEventListener("click", () => {
  void loadOwnedCases(false).catch((error) => showNotice(error.message, true));
});
document.querySelector("#refresh-owner-status").addEventListener("click", async () => {
  const button = document.querySelector("#refresh-owner-status");
  button.disabled = true;
  try {
    await refreshOwnedCaseStatus();
    showNotice("Состояние подтверждений обновлено.");
  } catch (error) {
    showNotice(`Не удалось обновить подтверждения. ${error.message}`, true);
  } finally {
    button.disabled = false;
  }
});
