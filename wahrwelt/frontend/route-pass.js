"use strict";

function routePassViewModel(state, isOwner) {
  const issued = Boolean(state?.issued && state?.pass);
  const ready = Boolean(state?.ready);
  let status = "Маршрут ожидает подтверждений. Пропуск еще не выпускался.";
  if (issued) {
    status = "Пропуск выпущен. Он остается доступен после истечения временных подтверждений.";
  } else if (ready) {
    status = isOwner
      ? "Маршрут готов. Выпустите пропуск для всех назначенных участников."
      : "Маршрут готов. Ожидаем выпуска пропуска владельцем дела.";
  }
  return {
    enabled: ready || issued,
    canIssue: Boolean(isOwner && ready && !issued),
    canDownload: issued,
    status,
  };
}

function buildRoutePassPNG(pass) {
  const kind = { import: "ВВОЗ", export: "ВЫВОЗ", service: "СЕРВИСНЫЙ РЕЙС" }[pass?.kind];
  const route = [pass?.origin, pass?.destination, pass?.case_id, pass?.route_slot_id, pass?.carrier_username, pass?.inspector_username];
  if (!kind || route.some((value) => typeof value !== "string" || !value.trim())) {
    throw new Error("Недостаточно данных пропуска для PNG.");
  }
  const issued = new Date(pass.issued_at);
  if (!pass.issued_at || Number.isNaN(issued.getTime())) {
    throw new Error("Неизвестно время выпуска пропуска.");
  }

  const canvas = document.createElement("canvas");
  canvas.width = 1224;
  canvas.height = 1584;
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("Браузер не смог подготовить PNG-пропуск.");
  const ink = "#173039";
  const green = "#17665b";
  const muted = "#536a70";
  ctx.fillStyle = "#eaf0ec";
  ctx.fillRect(0, 0, 1224, 1584);
  ctx.fillStyle = "#fffefa";
  ctx.fillRect(36, 36, 1152, 1512);
  ctx.fillStyle = ink;
  ctx.fillRect(36, 36, 1152, 405);

  function fitted(value, x, y, maxWidth, font, color) {
    ctx.font = font;
    ctx.fillStyle = color;
    let text = String(value).replace(/\s+/g, " ").trim();
    while (text.length > 1 && ctx.measureText(text).width > maxWidth) {
      text = text.slice(0, -2).trimEnd() + "…";
    }
    ctx.fillText(text, x, y);
  }
  function rule(y) {
    ctx.strokeStyle = "#c9d6d2";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(96, y);
    ctx.lineTo(1128, y);
    ctx.stroke();
  }

  fitted("WAHRWELT / ПРОПУСК", 96, 123, 1000, "600 29px sans-serif", "#ecf3ef");
  fitted("МАРШРУТ ГОТОВ", 96, 297, 1032, "700 83px sans-serif", "#fffefa");
  fitted("Состояние на момент выпуска", 96, 366, 1032, "400 29px sans-serif", "#c5d9d2");

  fitted("ОТПРАВЛЕНИЕ", 96, 524, 1000, "600 25px sans-serif", muted);
  fitted(pass.origin, 96, 630, 1032, "600 76px sans-serif", ink);
  rule(697);
  fitted("НАЗНАЧЕНИЕ", 96, 774, 1000, "600 25px sans-serif", muted);
  fitted(pass.destination, 96, 878, 1032, "600 76px sans-serif", ink);
  rule(935);
  fitted(`Дело: ${pass.case_id}`, 96, 1016, 1032, "500 35px monospace", ink);
  fitted(`Поставка: ${pass.route_slot_id}`, 96, 1080, 1032, "500 35px monospace", ink);
  fitted(`Операция: ${kind}`, 96, 1160, 1032, "600 30px sans-serif", ink);
  fitted(`Перевозчик: ${pass.carrier_username}`, 96, 1223, 1032, "500 31px sans-serif", ink);
  fitted(`Инспектор: ${pass.inspector_username}`, 96, 1275, 1032, "500 31px sans-serif", ink);

  const issuedUTC = issued.toISOString().replace("T", " ").slice(0, 16) + " UTC";
  ctx.fillStyle = "#e7f2ee";
  ctx.fillRect(96, 1340, 1032, 134);
  fitted(`Выпущен: ${issuedUTC}`, 124, 1401, 970, "600 31px sans-serif", green);
  fitted("Назначения не подтверждают личность подписанта.", 124, 1447, 970, "400 22px sans-serif", muted);

  const png = canvas.toDataURL("image/png");
  if (!png.startsWith("data:image/png;base64,")) throw new Error("Браузер не смог подготовить PNG-пропуск.");
  return png;
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { buildRoutePassPNG, routePassViewModel };
}
