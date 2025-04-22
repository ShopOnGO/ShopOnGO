let sseConnection = null;

async function connectToSSE(userID) {
    if (!userID) {
        console.error("Нет userID для подключения к SSE");
        return;
    }

    if (sseConnection) {
        console.log("⚠️ SSE уже подключено");
        return;
    }

    try {
        // Проверка: уже ли подключён этот userID на сервере
        const res = await fetch(`http://localhost:8079/sse/status/${userID}`);
        const data = await res.json();

        if (res.ok && data.connected) {
            console.log(`ℹ️ Пользователь ${userID} уже подключён к SSE (по данным сервера)`);
            return;
        }

        console.log("🔌 Подключаемся к SSE с userID:", userID);

        sseConnection = new EventSource(`http://localhost:8079/sse/${userID}`);

        sseConnection.onmessage = function (event) {
            console.log("🔔 Уведомление:", event.data);

            const container = document.getElementById("notifications");
            if (container) {
                const msg = document.createElement("div");
                msg.textContent = event.data;
                msg.className = "notification";
                container.appendChild(msg);
            }
        };

        sseConnection.onerror = function (err) {
            console.error("SSE ошибка:", err);
            sseConnection.close();
            sseConnection = null;
        };
    } catch (err) {
        console.error("Ошибка при проверке SSE статуса:", err);
    }
}
