// Render Cyrillic in the browser, then wrap the image for the existing EPS preview API.
function buildPreviewEPS(item) {
  const kind = { import: "ВВОЗ", export: "ВЫВОЗ", service: "СЕРВИСНЫЙ РЕЙС" }[item.kind];
  const origin = String(item.origin || "").trim();
  const destination = String(item.destination || "").trim();
  const caseID = String(item.case_id || "").trim();
  if (typeof kind !== "string" || !origin || !destination || !caseID) {
    throw new Error("Недостаточно данных дела для черновика.");
  }

  // The renderer produces a 612x792 page; double resolution keeps its text sharp.
  const width = 1224;
  const height = 1584;
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("Браузер не смог подготовить PNG-черновик.");

  const ink = "#173039";
  const muted = "#536a70";
  const green = "#167467";
  ctx.fillStyle = "#eaf0ec";
  ctx.fillRect(0, 0, width, height);
  ctx.fillStyle = "#fffefa";
  ctx.fillRect(36, 36, 1152, 1512);
  ctx.fillStyle = ink;
  ctx.fillRect(36, 36, 1152, 395);

  function line(x1, y1, x2, y2, color, thickness = 2) {
    ctx.strokeStyle = color;
    ctx.lineWidth = thickness;
    ctx.beginPath();
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.stroke();
  }

  function fitted(text, x, y, maxWidth, font, color) {
    ctx.font = font;
    ctx.fillStyle = color;
    const original = String(text).replace(/\s+/g, " ").trim();
    let value = original;
    if (ctx.measureText(value).width > maxWidth) {
      while (value && ctx.measureText(value + "…").width > maxWidth) {
        value = value.slice(0, -1);
      }
      value = value.trimEnd() + "…";
    }
    ctx.fillText(value, x, y);
  }

  ctx.font = "600 27px sans-serif";
  ctx.fillStyle = "#ecf3ef";
  ctx.fillText("WAHRWELT", 96, 115);
  ctx.font = kind.length > 10 ? "700 75px sans-serif" : "700 108px sans-serif";
  ctx.fillText(kind, 96, 334);

  ctx.fillStyle = muted;
  ctx.font = "600 25px sans-serif";
  ctx.fillText("ОТПРАВЛЕНИЕ", 96, 542);
  fitted(origin, 96, 651, 1020, "600 83px sans-serif", ink);
  ctx.fillStyle = green;
  ctx.beginPath();
  ctx.arc(120, 722, 9, 0, Math.PI * 2);
  ctx.fill();
  line(120, 740, 120, 820, green, 4);
  line(107, 810, 120, 824, green, 4);
  line(133, 810, 120, 824, green, 4);
  ctx.fillStyle = muted;
  ctx.font = "600 25px sans-serif";
  ctx.fillText("НАЗНАЧЕНИЕ", 96, 914);
  fitted(destination, 96, 1022, 1020, "600 83px sans-serif", ink);
  line(96, 1092, 1128, 1092, "#c9d6d2");

  ctx.fillStyle = muted;
  ctx.font = "600 25px sans-serif";
  ctx.fillText("НОМЕР ДЕЛА", 96, 1171);
  fitted(caseID, 96, 1238, 1020, "500 39px monospace", ink);

  ctx.fillStyle = "#f2ece1";
  ctx.fillRect(96, 1340, 1032, 124);
  ctx.fillStyle = "#70552a";
  ctx.font = "700 37px sans-serif";
  ctx.fillText("НЕ СОГЛАСОВАНО", 129, 1420);

  const jpeg = canvas.toDataURL("image/jpeg", 0.9);
  if (!jpeg.startsWith("data:image/jpeg;base64,")) {
    throw new Error("Браузер не смог подготовить PNG-черновик.");
  }
  const binary = atob(jpeg.slice("data:image/jpeg;base64,".length));
  let hex = "";
  for (let i = 0; i < binary.length; i++) {
    hex += binary.charCodeAt(i).toString(16).padStart(2, "0");
  }
  const eps = [
    "%!PS-Adobe-3.0 EPSF-3.0",
    "%%BoundingBox: 0 0 612 792",
    "%%LanguageLevel: 2",
    "gsave",
    "612 792 scale",
    "/DeviceRGB setcolorspace",
    `<< /ImageType 1 /Width ${width} /Height ${height} /BitsPerComponent 8`,
    `/Decode [0 1 0 1 0 1] /ImageMatrix [${width} 0 0 -${height} 0 ${height}]`,
    "/DataSource currentfile /ASCIIHexDecode filter /DCTDecode filter >> image",
    hex,
    ">",
    "grestore",
    "showpage",
    "%%EOF",
    "",
  ].join("\n");
  if (eps.length > 950000) throw new Error("Черновик слишком велик для renderer.");
  return eps;
}
