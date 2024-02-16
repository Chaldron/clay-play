document.addEventListener("htmx:beforeSwap", (e) => {
    if (
        e.detail.xhr.status === 500
        || e.detail.xhr.status === 400
        || e.detail.xhr.status === 401
        || e.detail.xhr.status === 403
    ) {
        e.detail.shouldSwap = true
        e.detail.isError = false
    }
})

function formatTime(t) {
    return dayjs(t).format("ddd, MMM DD h:mm A")
}

function formFormatTime(t) {
    return dayjs(t).format("YYYY-MM-DDTHH:mm")
}
