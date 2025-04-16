function connectToSSE(userID) {
    if (!userID) {
        console.error("Нет userID для подключения к SSE");
        return;
    }

    console.log("🔌 Подключаемся к SSE с userID:", userID); // ← лог перед подключением

    const eventSource = new EventSource(`http://localhost:8079/sse/${userID}`);

    eventSource.onmessage = function (event) {
        console.log("🔔 Уведомление:", event.data);

        const container = document.getElementById("notifications");
        if (container) {
            const msg = document.createElement("div");
            msg.textContent = event.data;
            msg.className = "notification";
            container.appendChild(msg);
        }
    };

    eventSource.onerror = function (err) {
        console.error("SSE ошибка:", err);
        eventSource.close();
    };
}
