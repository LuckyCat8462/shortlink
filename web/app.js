(function () {
  const KEY_STORAGE = "shortlink.apiKey";
  const SESSION_STORAGE = "shortlink.assistantSession";

  const el = {
    form: document.getElementById("create-form"),
    apiKey: document.getElementById("api-key"),
    longUrl: document.getElementById("long-url"),
    expireDays: document.getElementById("expire-days"),
    sceneId: document.getElementById("scene-id"),
    formError: document.getElementById("form-error"),
    btnCreate: document.getElementById("btn-create"),
    resultPanel: document.getElementById("result-panel"),
    shortUrl: document.getElementById("short-url"),
    shortCode: document.getElementById("short-code"),
    created: document.getElementById("created"),
    expireAt: document.getElementById("expire-at"),
    btnCopy: document.getElementById("btn-copy"),
    btnOpen: document.getElementById("btn-open"),
    btnStats: document.getElementById("btn-stats"),
    copyHint: document.getElementById("copy-hint"),
    statsPanel: document.getElementById("stats-panel"),
    statPv: document.getElementById("stat-pv"),
    statUv: document.getElementById("stat-uv"),
    statLa: document.getElementById("stat-la"),
    statsError: document.getElementById("stats-error"),
    chatLog: document.getElementById("chat-log"),
    chatForm: document.getElementById("chat-form"),
    chatInput: document.getElementById("chat-input"),
    btnChat: document.getElementById("btn-chat"),
    chatError: document.getElementById("chat-error"),
  };

  let currentCode = "";
  let sessionId = localStorage.getItem(SESSION_STORAGE) || "";

  const savedKey = localStorage.getItem(KEY_STORAGE);
  if (savedKey) {
    el.apiKey.value = savedKey;
  }

  function showError(node, text) {
    if (!text) {
      node.hidden = true;
      node.textContent = "";
      return;
    }
    node.hidden = false;
    node.textContent = text;
  }

  function apiKey() {
    return el.apiKey.value.trim();
  }

  async function parseError(res) {
    try {
      const data = await res.json();
      if (data && data.msg) return data.msg;
    } catch (_) {}
    if (res.status === 401) return "未授权：请检查 API Key";
    if (res.status === 429) return "请求过于频繁，请稍后再试";
    return "请求失败 (" + res.status + ")";
  }

  el.form.addEventListener("submit", async function (e) {
    e.preventDefault();
    showError(el.formError, "");
    const key = apiKey();
    if (!key) {
      showError(el.formError, "请填写 API Key");
      return;
    }
    localStorage.setItem(KEY_STORAGE, key);

    const longUrl = el.longUrl.value.trim();
    const expireDays = Number(el.expireDays.value);
    const sceneId = el.sceneId.value.trim();

    const body = { longUrl: longUrl };
    if (!Number.isNaN(expireDays)) {
      body.expireDays = expireDays;
    }
    if (sceneId) {
      body.sceneId = sceneId;
    }

    el.btnCreate.disabled = true;
    try {
      const res = await fetch("/api/v1/links", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": key,
        },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        showError(el.formError, await parseError(res));
        return;
      }
      const data = await res.json();
      currentCode = data.shortCode || "";
      el.shortUrl.textContent = data.shortUrl || "";
      el.shortUrl.href = data.shortUrl || "#";
      el.btnOpen.href = data.shortUrl || "#";
      el.shortCode.textContent = data.shortCode || "—";
      el.created.textContent = data.created ? "是 (created=true)" : "否 (幂等命中)";
      el.expireAt.textContent = data.expireAt || "永不过期";
      el.resultPanel.hidden = false;
      el.statsPanel.hidden = false;
      showError(el.copyHint, "");
      await refreshStats();
    } catch (err) {
      showError(el.formError, "网络错误：" + (err && err.message ? err.message : String(err)));
    } finally {
      el.btnCreate.disabled = false;
    }
  });

  el.btnCopy.addEventListener("click", async function () {
    const url = el.shortUrl.textContent;
    if (!url) return;
    try {
      await navigator.clipboard.writeText(url);
      showError(el.copyHint, "已复制");
      setTimeout(function () {
        showError(el.copyHint, "");
      }, 1500);
    } catch (_) {
      showError(el.formError, "复制失败，请手动选择短链");
    }
  });

  el.btnStats.addEventListener("click", function () {
    refreshStats();
  });

  async function refreshStats() {
    showError(el.statsError, "");
    if (!currentCode) {
      showError(el.statsError, "暂无短码");
      return;
    }
    const key = apiKey();
    if (!key) {
      showError(el.statsError, "请填写 API Key");
      return;
    }
    try {
      const res = await fetch("/api/v1/links/" + encodeURIComponent(currentCode) + "/stats", {
        headers: { "X-API-Key": key },
      });
      if (!res.ok) {
        showError(el.statsError, await parseError(res));
        return;
      }
      const data = await res.json();
      el.statPv.textContent = String(data.pv != null ? data.pv : 0);
      el.statUv.textContent = String(data.uv != null ? data.uv : 0);
      el.statLa.textContent = data.lastClickAt || "—";
    } catch (err) {
      showError(el.statsError, "网络错误：" + (err && err.message ? err.message : String(err)));
    }
  }

  function appendBubble(role, text, pending) {
    const div = document.createElement("div");
    div.className = "chat-bubble " + role + (pending ? " pending" : "");
    div.textContent = text || "";
    el.chatLog.appendChild(div);
    el.chatLog.scrollTop = el.chatLog.scrollHeight;
    return div;
  }

  async function readSSE(res, handlers) {
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buf = "";
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      let sep;
      while ((sep = buf.indexOf("\n\n")) >= 0) {
        const chunk = buf.slice(0, sep);
        buf = buf.slice(sep + 2);
        let event = "message";
        let data = "";
        chunk.split("\n").forEach(function (line) {
          if (line.indexOf("event:") === 0) event = line.slice(6).trim();
          else if (line.indexOf("data:") === 0) data += line.slice(5).trim();
        });
        if (!data) continue;
        let payload = data;
        try {
          payload = JSON.parse(data);
        } catch (_) {}
        if (handlers[event]) handlers[event](payload);
      }
    }
  }

  el.chatForm.addEventListener("submit", async function (e) {
    e.preventDefault();
    showError(el.chatError, "");
    const key = apiKey();
    if (!key) {
      showError(el.chatError, "请先填写上方 API Key");
      return;
    }
    localStorage.setItem(KEY_STORAGE, key);
    const message = el.chatInput.value.trim();
    if (!message) return;

    appendBubble("user", message, false);
    el.chatInput.value = "";
    const bot = appendBubble("assistant", "思考中…", true);
    el.btnChat.disabled = true;

    try {
      const res = await fetch("/api/v1/assistant/chat", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": key,
          Accept: "text/event-stream",
        },
        body: JSON.stringify({ sessionId: sessionId, message: message }),
      });
      if (!res.ok) {
        const msg = await parseError(res);
        bot.classList.remove("pending");
        bot.textContent = msg;
        showError(el.chatError, msg);
        return;
      }

      let gotDelta = false;
      await readSSE(res, {
        session: function (p) {
          if (p && p.sessionId) {
            sessionId = p.sessionId;
            localStorage.setItem(SESSION_STORAGE, sessionId);
          }
        },
        status: function () {},
        delta: function (p) {
          const text = p && p.text != null ? p.text : String(p || "");
          if (!gotDelta) {
            bot.textContent = "";
            bot.classList.remove("pending");
            gotDelta = true;
          }
          bot.textContent += text;
          el.chatLog.scrollTop = el.chatLog.scrollHeight;
        },
        error: function (p) {
          const msg = (p && p.message) || "助手错误";
          bot.classList.remove("pending");
          if (!gotDelta) bot.textContent = msg;
          showError(el.chatError, msg);
        },
        done: function () {
          bot.classList.remove("pending");
          if (!gotDelta && !bot.textContent) bot.textContent = "（无回复）";
        },
      });
    } catch (err) {
      bot.classList.remove("pending");
      bot.textContent = "网络错误";
      showError(el.chatError, "网络错误：" + (err && err.message ? err.message : String(err)));
    } finally {
      el.btnChat.disabled = false;
    }
  });
})();
