import { Events, Window } from "@wailsio/runtime";

const PET_PROFILES = Object.freeze({
    neko: Object.freeze({
        id: "neko",
        label: "CyberNeko",
        menuId: "neko-menu",
        greetingLine: "喵，我先在这里巡逻啦",
        doubleClickLine: "嘿嘿，再点一下嘛",
        initialSpeechDelay: 1600,
        speechLines: Object.freeze([
            "主人，摸摸头嘛",
            "不要不理我嘛",
            "我今天也超乖",
            "陪我玩一会儿",
            "喵，想你啦",
            "可以抱抱我吗",
            "我会一直陪着你",
            "给我一点小鱼干",
            "别工作太久啦",
            "看我，看我嘛",
        ]),
    }),
    momo: Object.freeze({
        id: "momo",
        label: "CyberMomo",
        menuId: "momo-menu",
        greetingLine: "我也来啦，贴贴",
        doubleClickLine: "再戳我就撒娇给你看",
        initialSpeechDelay: 2800,
        speechLines: Object.freeze([
            "要不要一起摸鱼呀",
            "我在旁边陪你哦",
            "今天也要被夸夸",
            "贴贴一下就有精神",
            "你的鼠标跑得好快",
            "等你忙完抱抱我",
            "我把好运放这里",
            "小点心分我一口",
            "休息三分钟嘛",
            "我超会撒娇的",
        ]),
    }),
});

const requestedPetId = new URLSearchParams(window.location.search).get("pet") || "neko";
const petProfile = PET_PROFILES[requestedPetId] || PET_PROFILES.neko;
const petElement = document.getElementById("neko");
const speechBubble = document.getElementById("neko-speech");

document.title = petProfile.label;
document.documentElement.dataset.pet = petProfile.id;
document.body.dataset.pet = petProfile.id;
petElement.dataset.pet = petProfile.id;
petElement.style.setProperty("--custom-contextmenu", petProfile.menuId);
petElement.style.setProperty("--custom-contextmenu-data", petProfile.id);
petElement.setAttribute("aria-label", petProfile.label + " desktop pet");

let autoMovePausedUntil = 0;
const manualDrag = {
    active: false,
    ready: false,
    pointerId: null,
    startScreenX: 0,
    startScreenY: 0,
    startWindowX: 0,
    startWindowY: 0,
    nextX: 0,
    nextY: 0,
    frame: 0,
};

const PetState = Object.freeze({
    Idle: "idle",
    Walking: "walking",
});

const SPEECH_VISIBLE_MS = 3000;
const SPEECH_MIN_DELAY_MS = 7000;
const SPEECH_MAX_DELAY_MS = 15000;

function pauseAutoMove(duration = 1200) {
    autoMovePausedUntil = Math.max(autoMovePausedUntil, Date.now() + duration);
}

window.__cyberNekoSetPosition = (x, y) => {
    if (Date.now() < autoMovePausedUntil) {
        return;
    }
    Window.SetPosition(x, y);
};

let currentState = PetState.Idle;
let currentFrame = 0;
let lastSpeechIndex = -1;
let speechHideTimer = 0;
let speechNextTimer = 0;

function randomBetween(min, max) {
    return min + Math.floor(Math.random() * (max - min + 1));
}

function pickSpeechLine() {
    const lines = petProfile.speechLines;
    if (lines.length === 1) {
        lastSpeechIndex = 0;
        return lines[0];
    }

    let nextIndex = lastSpeechIndex;
    while (nextIndex === lastSpeechIndex) {
        nextIndex = Math.floor(Math.random() * lines.length);
    }
    lastSpeechIndex = nextIndex;
    return lines[nextIndex];
}

function showSpeechBubble(text = pickSpeechLine()) {
    window.clearTimeout(speechHideTimer);
    speechBubble.textContent = text;
    speechBubble.dataset.visible = "true";
    speechHideTimer = window.setTimeout(() => {
        speechBubble.dataset.visible = "false";
    }, SPEECH_VISIBLE_MS);
}

function scheduleNextSpeechBubble() {
    window.clearTimeout(speechNextTimer);
    speechNextTimer = window.setTimeout(() => {
        showSpeechBubble();
        scheduleNextSpeechBubble();
    }, randomBetween(SPEECH_MIN_DELAY_MS, SPEECH_MAX_DELAY_MS));
}

function setPetState(nextState) {
    if (!Object.values(PetState).includes(nextState)) {
        return;
    }

    currentState = nextState;
    currentFrame = 0;
    petElement.dataset.state = nextState;
    petElement.dataset.frame = String(currentFrame);
}

function setPetDirection(nextDirection) {
    if (nextDirection === "left" || nextDirection === "right") {
        petElement.dataset.direction = nextDirection;
    }
}

window.__cyberNekoSetState = setPetState;
window.__cyberNekoSetDirection = setPetDirection;

// 第一阶段的动画循环。
// 后续替换为真实 sprite sheet 时，只需要在这里根据 currentState/currentFrame 切换 background-position。
window.setInterval(() => {
    const frameCount = currentState === PetState.Walking ? 4 : 2;
    currentFrame = (currentFrame + 1) % frameCount;
    petElement.dataset.frame = String(currentFrame);
}, 150);

// 接收 Go 侧右键菜单发出的全局状态切换事件。行走器自己的状态会通过 ExecJS 只发给对应窗口。
Events.On("pet:state", (event) => {
    setPetState(event.data);
});

Events.On("pet:direction", (event) => {
    setPetDirection(event.data);
});

function flushManualDragPosition() {
    manualDrag.frame = 0;
    if (!manualDrag.active || !manualDrag.ready) {
        return;
    }
    Window.SetPosition(manualDrag.nextX, manualDrag.nextY);
}

function queueManualDragPosition(x, y) {
    manualDrag.nextX = x;
    manualDrag.nextY = y;
    if (manualDrag.frame !== 0) {
        return;
    }
    manualDrag.frame = window.requestAnimationFrame(flushManualDragPosition);
}

async function startManualDrag(event) {
    if (event.button !== 0) {
        return;
    }

    event.preventDefault();
    pauseAutoMove(6000);
    manualDrag.active = true;
    manualDrag.ready = false;
    manualDrag.pointerId = event.pointerId;
    manualDrag.startScreenX = event.screenX;
    manualDrag.startScreenY = event.screenY;

    try {
        petElement.setPointerCapture(event.pointerId);
    } catch {
        // Pointer capture may be unavailable in older WebView runtimes; dragging still works while hovered.
    }

    const position = await Window.Position();
    if (!manualDrag.active || manualDrag.pointerId !== event.pointerId) {
        return;
    }

    manualDrag.startWindowX = position.x;
    manualDrag.startWindowY = position.y;
    manualDrag.nextX = position.x;
    manualDrag.nextY = position.y;
    manualDrag.ready = true;
}

function updateManualDrag(event) {
    if (!manualDrag.active || manualDrag.pointerId !== event.pointerId) {
        return;
    }

    pauseAutoMove(6000);
    if (!manualDrag.ready) {
        return;
    }

    const nextX = Math.round(manualDrag.startWindowX + event.screenX - manualDrag.startScreenX);
    const nextY = Math.round(manualDrag.startWindowY + event.screenY - manualDrag.startScreenY);
    queueManualDragPosition(nextX, nextY);
}

function stopManualDrag(event) {
    if (manualDrag.pointerId !== null && event.pointerId !== manualDrag.pointerId) {
        return;
    }

    pauseAutoMove(1600);
    manualDrag.active = false;
    manualDrag.ready = false;
    manualDrag.pointerId = null;

    if (manualDrag.frame !== 0) {
        window.cancelAnimationFrame(manualDrag.frame);
        manualDrag.frame = 0;
    }
}

petElement.addEventListener("pointerdown", startManualDrag);
petElement.addEventListener("pointermove", updateManualDrag);

for (const eventName of ["pointerup", "pointercancel", "lostpointercapture"]) {
    petElement.addEventListener(eventName, stopManualDrag);
    window.addEventListener(eventName, stopManualDrag);
}

window.addEventListener("blur", () => {
    pauseAutoMove();
    manualDrag.active = false;
    manualDrag.ready = false;
    manualDrag.pointerId = null;
});

// 双击主体也切换一次状态，便于开发时不用每次打开右键菜单。
petElement.addEventListener("dblclick", () => {
    setPetState(currentState === PetState.Idle ? PetState.Walking : PetState.Idle);
    showSpeechBubble(petProfile.doubleClickLine);
});

setPetState(PetState.Idle);
setPetDirection("right");
window.setTimeout(() => showSpeechBubble(petProfile.greetingLine), petProfile.initialSpeechDelay);
scheduleNextSpeechBubble();
