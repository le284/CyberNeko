import { Call, Events, Window } from "@wailsio/runtime";

import cat30Url from "../../cat/3-0-transparent.webp";
import cat31Url from "../../cat/3-1.png";
import cat32Url from "../../cat/3-2.png";
import cat33Url from "../../cat/3-3.png";
import cat34Url from "../../cat/3-4.png";
import cat35Url from "../../cat/3-5-transparent.webp";

const SERVICE_PREFIX = "main.PetService.";
const PET_IMAGES_STORAGE_KEY = "cyberneko.petImages.v1";
const CUSTOM_IMAGE_MAX_SIDE = 360;
const CUSTOM_GIF_MAX_BYTES = 4 * 1024 * 1024;
const BACKGROUND_ALPHA_THRESHOLD = 238;
const IS_MAC_PLATFORM = /mac|iphone|ipad|ipod/i.test(navigator.userAgentData?.platform || navigator.platform || "");

const DEFAULT_SHORTCUTS = Object.freeze({
    idle: "Ctrl+Alt+1",
    edgeWander: "Ctrl+Alt+2",
    followMouse: "Ctrl+Alt+3",
    cycle: "Ctrl+Alt+Space",
});

const SHORTCUT_ACTIONS = Object.freeze([
    Object.freeze({ id: "idle", label: "原地待机" }),
    Object.freeze({ id: "edgeWander", label: "沿边缘巡游" }),
    Object.freeze({ id: "followMouse", label: "跟随鼠标" }),
    Object.freeze({ id: "cycle", label: "循环切换" }),
]);

const PetAction = Object.freeze({
    Idle: "idle",
    Walking: "walking",
    Wave: "wave",
    HeadTilt: "headTilt",
    Stretch: "stretch",
    Blink: "blink",
    Cute: "cute",
});

const PET_ACTIONS = Object.freeze([
    Object.freeze({ id: PetAction.Wave, label: "招手", duration: 1800 }),
    Object.freeze({ id: PetAction.HeadTilt, label: "歪头", duration: 2200 }),
    Object.freeze({ id: PetAction.Stretch, label: "伸懒腰", duration: 2000 }),
    Object.freeze({ id: PetAction.Blink, label: "眨眼睛", duration: 900 }),
    Object.freeze({ id: PetAction.Cute, label: "卖萌", duration: 2400 }),
]);

const PET_ACTION_IMAGES = Object.freeze({
    [PetAction.Idle]: cat31Url,
    [PetAction.Walking]: cat30Url,
    [PetAction.Wave]: cat31Url,
    [PetAction.HeadTilt]: cat32Url,
    [PetAction.Stretch]: cat33Url,
    [PetAction.Blink]: cat34Url,
    [PetAction.Cute]: cat35Url,
});

const DEFAULT_PET_ACTION = PetAction.Idle;
const MAX_PET_SLOTS = 1;
const transparentActionImageUrls = new Map();
const transparentActionImagePromises = new Map();

const PET_PROFILES = Object.freeze({
    neko: Object.freeze({
        id: "neko",
        label: "CyberNeko",
        menuId: "neko-menu",
        greetingLine: "喵，我先在这里巡逻啦",
        doubleClickLine: "嘿嘿，再点一下嘛",
        theme: Object.freeze({
            "--pet-fur": "#343741",
            "--pet-outline": "#17191f",
            "--pet-ear": "#ff8ea3",
            "--pet-eye": "#6ff3d6",
            "--pet-eye-glow": "rgba(111, 243, 214, 0.75)",
            "--pet-bubble-text": "#20242c",
            "--pet-bubble-border": "rgba(23, 25, 31, 0.9)",
        }),
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
        theme: Object.freeze({
            "--pet-fur": "#f2a7c3",
            "--pet-outline": "#21151d",
            "--pet-ear": "#ffd166",
            "--pet-eye": "#8ee86f",
            "--pet-eye-glow": "rgba(142, 232, 111, 0.78)",
            "--pet-bubble-text": "#2b1622",
            "--pet-bubble-border": "rgba(66, 27, 42, 0.92)",
        }),
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
    sora: Object.freeze({
        id: "sora",
        label: "CyberSora",
        menuId: "sora-menu",
        greetingLine: "我找到一块好边缘",
        doubleClickLine: "我转身给你看",
        theme: Object.freeze({
            "--pet-fur": "#7ab6ff",
            "--pet-outline": "#172036",
            "--pet-ear": "#ffc0dd",
            "--pet-eye": "#fff275",
            "--pet-eye-glow": "rgba(255, 242, 117, 0.76)",
            "--pet-bubble-text": "#152033",
            "--pet-bubble-border": "rgba(23, 32, 54, 0.9)",
        }),
        speechLines: Object.freeze([
            "我飞快但不捣乱",
            "边缘巡逻完成",
            "今天也守着你",
            "鼠标先生等等我",
            "给你一点好运",
        ]),
    }),
    luna: Object.freeze({
        id: "luna",
        label: "CyberLuna",
        menuId: "luna-menu",
        greetingLine: "月光小队到位",
        doubleClickLine: "我眨眼啦",
        theme: Object.freeze({
            "--pet-fur": "#7d6df2",
            "--pet-outline": "#201944",
            "--pet-ear": "#94f2d4",
            "--pet-eye": "#ffcf5a",
            "--pet-eye-glow": "rgba(255, 207, 90, 0.72)",
            "--pet-bubble-text": "#211944",
            "--pet-bubble-border": "rgba(32, 25, 68, 0.9)",
        }),
        speechLines: Object.freeze([
            "悄悄陪你加班",
            "我会乖乖站好",
            "你的窗口我守住了",
            "要记得喝水呀",
            "今晚也很可爱",
        ]),
    }),
    mika: Object.freeze({
        id: "mika",
        label: "CyberMika",
        menuId: "mika-menu",
        greetingLine: "新的巡游路线开始",
        doubleClickLine: "戳到了我的开心开关",
        theme: Object.freeze({
            "--pet-fur": "#ffd15c",
            "--pet-outline": "#30230c",
            "--pet-ear": "#ff9aa2",
            "--pet-eye": "#56e5c4",
            "--pet-eye-glow": "rgba(86, 229, 196, 0.75)",
            "--pet-bubble-text": "#30230c",
            "--pet-bubble-border": "rgba(48, 35, 12, 0.9)",
        }),
        speechLines: Object.freeze([
            "想被夸一下",
            "我在这里发光",
            "工作顺顺利利",
            "小爪子在巡逻",
            "再忙也要休息",
        ]),
    }),
});

const query = new URLSearchParams(window.location.search);
const currentView = query.get("view") === "settings" ? "settings" : "pet";
document.body.dataset.view = currentView;

if (currentView === "settings") {
    initSettingsWindow();
} else {
    initPetWindow();
}

function callService(method, ...args) {
    return Call.ByName(SERVICE_PREFIX + method, ...args);
}

function normalizeSettings(rawSettings) {
    const petCount = Number(rawSettings?.petCount ?? rawSettings?.PetCount ?? 1);
    const maxPets = Number(rawSettings?.maxPets ?? rawSettings?.MaxPets ?? 6);
    return {
        petCount: clampNumber(petCount, 1, maxPets || 6),
        maxPets: clampNumber(maxPets || 6, 1, 12),
        shortcuts: normalizeShortcutSettings(rawSettings?.shortcuts ?? rawSettings?.Shortcuts),
    };
}

function normalizeShortcutSettings(rawShortcuts) {
    const source = rawShortcuts && typeof rawShortcuts === "object" ? rawShortcuts : {};
    return {
        idle: normalizeShortcutText(source.idle ?? source.Idle, DEFAULT_SHORTCUTS.idle),
        edgeWander: normalizeShortcutText(source.edgeWander ?? source.EdgeWander, DEFAULT_SHORTCUTS.edgeWander),
        followMouse: normalizeShortcutText(source.followMouse ?? source.FollowMouse, DEFAULT_SHORTCUTS.followMouse),
        cycle: normalizeShortcutText(source.cycle ?? source.Cycle, DEFAULT_SHORTCUTS.cycle),
    };
}

function normalizeShortcutText(value, fallback) {
    const text = String(value || "").trim();
    return text || fallback;
}

function clampNumber(value, min, max) {
    const numericValue = Number.isFinite(value) ? value : min;
    return Math.min(max, Math.max(min, Math.round(numericValue)));
}

function readPetImages() {
    try {
        const parsed = JSON.parse(window.localStorage.getItem(PET_IMAGES_STORAGE_KEY) || "{}");
        return parsed && typeof parsed === "object" ? parsed : {};
    } catch {
        return {};
    }
}

function writePetImages(images) {
    window.localStorage.setItem(PET_IMAGES_STORAGE_KEY, JSON.stringify(images));
}

function getStoredPetImage(slot) {
    return readPetImages()[String(slot)] || "";
}

function actionImageForSlot(slot, actionId = DEFAULT_PET_ACTION) {
    return PET_ACTION_IMAGES[actionId] || PET_ACTION_IMAGES[DEFAULT_PET_ACTION] || "";
}

function setStoredPetImage(slot, dataUrl) {
    const images = readPetImages();
    if (dataUrl) {
        images[String(slot)] = dataUrl;
    } else {
        delete images[String(slot)];
    }
    writePetImages(images);
}

function initPetWindow() {
    const requestedPetId = query.get("pet") || "neko";
    const petSlot = clampNumber(Number(query.get("slot") || 0), 0, 99);
    const petProfile = PET_PROFILES[requestedPetId] || PET_PROFILES.neko;
    const petElement = document.getElementById("neko");
    const actionImageElement = document.getElementById("pet-action-image");
    const customImageElement = document.getElementById("pet-custom-image");
    const speechBubble = document.getElementById("neko-speech");

    petElement.dataset.actionImage = "false";
    actionImageElement.addEventListener("load", () => {
        petElement.dataset.actionImage = actionImageElement.currentSrc ? "true" : "false";
    });
    actionImageElement.addEventListener("error", () => {
        actionImageElement.removeAttribute("src");
        petElement.dataset.actionImage = "false";
    });

    document.title = petProfile.label;
    document.documentElement.dataset.pet = petProfile.id;
    document.body.dataset.pet = petProfile.id;
    petElement.dataset.pet = petProfile.id;
    petElement.dataset.slot = String(petSlot);
    petElement.style.setProperty("--custom-contextmenu", petProfile.menuId);
    petElement.style.setProperty("--custom-contextmenu-data", petProfile.id);
    petElement.setAttribute("aria-label", petProfile.label + " desktop pet");

    for (const [key, value] of Object.entries(petProfile.theme)) {
        petElement.style.setProperty(key, value);
        speechBubble.style.setProperty(key, value);
    }

    applyCustomImage();
    Events.On("pet:visuals", () => applyCustomImage());

    let autoMovePausedUntil = 0;
    const manualDrag = {
        active: false,
        ready: false,
        moved: false,
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
    let currentAction = DEFAULT_PET_ACTION;
    let currentFrame = 0;
    let lastSpeechIndex = -1;
    let actionTimer = 0;
    let speechHideTimer = 0;
    let speechNextTimer = 0;

    function applyCustomImage() {
        const imageData = getStoredPetImage(petSlot);
        if (!imageData) {
            customImageElement.removeAttribute("src");
            petElement.dataset.customImage = "false";
            return;
        }

        customImageElement.src = imageData;
        petElement.dataset.customImage = "true";
    }

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

    async function setPetAction(actionId) {
        const imageUrl = PET_ACTION_IMAGES[actionId];
        if (!imageUrl) {
            return;
        }

        currentAction = actionId;
        petElement.dataset.action = actionId;
        const actionImageUrl = await transparentActionImageUrl(imageUrl);
        if (currentAction !== actionId) {
            return;
        }
        actionImageElement.src = actionImageUrl;
    }

    function playPetAction(actionId) {
        const action = PET_ACTIONS.find((item) => item.id === actionId);
        if (!action) {
            return;
        }

        window.clearTimeout(actionTimer);
        setPetAction(action.id);
        showSpeechBubble(action.label);

        if (action.id !== DEFAULT_PET_ACTION) {
            actionTimer = window.setTimeout(() => {
                setPetAction(currentState === PetState.Walking ? PetAction.Walking : DEFAULT_PET_ACTION);
            }, action.duration);
        }
    }

    function setPetState(nextState) {
        if (!Object.values(PetState).includes(nextState)) {
            return;
        }

        currentState = nextState;
        currentFrame = 0;
        petElement.dataset.state = nextState;
        petElement.dataset.frame = String(currentFrame);

        if (nextState === PetState.Walking) {
            setPetAction(PetAction.Walking);
            return;
        }
        if (currentAction === PetAction.Walking) {
            setPetAction(DEFAULT_PET_ACTION);
        }
    }

    function setPetDirection(nextDirection) {
        if (nextDirection === "left" || nextDirection === "right") {
            petElement.dataset.direction = nextDirection;
        }
    }

    function setPetEdge(nextEdge) {
        if (["top", "bottom", "free"].includes(nextEdge)) {
            petElement.dataset.edge = nextEdge;
        }
    }

    window.__cyberNekoSetState = setPetState;
    window.__cyberNekoSetDirection = setPetDirection;
    window.__cyberNekoSetEdge = setPetEdge;

    window.setInterval(() => {
        const frameCount = currentState === PetState.Walking ? 4 : 2;
        currentFrame = (currentFrame + 1) % frameCount;
        petElement.dataset.frame = String(currentFrame);
    }, 150);

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
        setPetEdge("free");
        pauseAutoMove(6000);
        manualDrag.active = true;
        manualDrag.ready = false;
        manualDrag.moved = false;
        manualDrag.pointerId = event.pointerId;
        manualDrag.startScreenX = event.screenX;
        manualDrag.startScreenY = event.screenY;

        try {
            petElement.setPointerCapture(event.pointerId);
        } catch {
            // Older WebView runtimes may not support pointer capture.
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
        if (Math.abs(event.screenX - manualDrag.startScreenX) > 3 || Math.abs(event.screenY - manualDrag.startScreenY) > 3) {
            manualDrag.moved = true;
        }
        queueManualDragPosition(nextX, nextY);
    }

    function stopManualDrag(event) {
        const wasDragging = manualDrag.active && manualDrag.moved;
        if (manualDrag.pointerId !== null && event.pointerId !== manualDrag.pointerId) {
            return;
        }

        pauseAutoMove(1600);
        manualDrag.active = false;
        manualDrag.ready = false;
        manualDrag.moved = false;
        manualDrag.pointerId = null;

        if (manualDrag.frame !== 0) {
            window.cancelAnimationFrame(manualDrag.frame);
            manualDrag.frame = 0;
        }

        if (wasDragging) {
            playPetAction(PetAction.Stretch);
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
        manualDrag.moved = false;
        manualDrag.pointerId = null;
    });

    petElement.addEventListener("dblclick", () => {
        playPetAction(PetAction.Cute);
    });

    petElement.addEventListener("click", (event) => {
        if (event.detail === 1 && currentAction !== PetAction.Cute) {
            playPetAction(PetAction.Blink);
        }
    });

    setPetState(PetState.Idle);
    setPetDirection("right");
    setPetEdge("free");
    setPetAction(DEFAULT_PET_ACTION);
    window.setTimeout(() => showSpeechBubble(petProfile.greetingLine), 1600 + petSlot * 450);
    scheduleNextSpeechBubble();
}

async function initSettingsWindow() {
    document.title = "CyberNeko 设置";

    const rangeInput = document.getElementById("pet-count-range");
    const numberInput = document.getElementById("pet-count-number");
    const output = document.getElementById("pet-count-output");
    const saveButton = document.getElementById("pet-count-save");
    const closeButton = document.getElementById("settings-close");
    const imageList = document.getElementById("pet-image-list");
    const shortcutList = document.getElementById("shortcut-list");
    const shortcutSaveButton = document.getElementById("shortcut-save");
    const status = document.getElementById("settings-status");

    let settings = normalizeSettings(await safeGetSettings());
    let shortcutDraft = { ...settings.shortcuts };
    let capturingShortcutAction = "";
    applySettingsToControls(settings);
    renderShortcutRows();
    renderPetImageSlots(settings);

    closeButton.addEventListener("click", () => Window.Hide());
    saveButton.addEventListener("click", savePetCount);
    shortcutSaveButton.addEventListener("click", saveShortcuts);
    document.addEventListener("keydown", captureShortcutFromKeyboard, true);
    rangeInput.addEventListener("input", () => syncCountInputs(rangeInput.value));
    numberInput.addEventListener("input", () => syncCountInputs(numberInput.value));

    Events.On("pet:settings", (event) => {
        settings = normalizeSettings(event.data);
        shortcutDraft = { ...settings.shortcuts };
        applySettingsToControls(settings);
        renderShortcutRows();
        renderPetImageSlots(settings);
    });

    async function safeGetSettings() {
        try {
            return await callService("GetSettings");
        } catch (error) {
            showStatus("设置服务暂时不可用", true);
            console.error(error);
            return { petCount: 1, maxPets: 6, shortcuts: DEFAULT_SHORTCUTS };
        }
    }

    function applySettingsToControls(nextSettings) {
        rangeInput.max = String(nextSettings.maxPets);
        numberInput.max = String(nextSettings.maxPets);
        syncCountInputs(nextSettings.petCount);
    }

    function syncCountInputs(value) {
        const count = clampNumber(Number(value), 1, settings.maxPets);
        rangeInput.value = String(count);
        numberInput.value = String(count);
        output.value = String(count);
        output.textContent = String(count);
    }

    function renderShortcutRows() {
        shortcutList.replaceChildren();
        for (const action of SHORTCUT_ACTIONS) {
            const row = document.createElement("article");
            row.className = "shortcut-row";
            row.dataset.capturing = String(capturingShortcutAction === action.id);

            const title = document.createElement("span");
            title.className = "shortcut-title";
            title.textContent = action.label;

            const input = document.createElement("input");
            input.className = "shortcut-input";
            input.readOnly = true;
            input.value = capturingShortcutAction === action.id ? "按下组合键" : shortcutDraft[action.id];
            input.setAttribute("aria-label", action.label + "快捷键");
            input.addEventListener("focus", () => startShortcutCapture(action.id));
            input.addEventListener("click", () => startShortcutCapture(action.id));

            const recordButton = document.createElement("button");
            recordButton.className = "secondary-button shortcut-button";
            recordButton.type = "button";
            recordButton.textContent = capturingShortcutAction === action.id ? "录制中" : "录制";
            recordButton.addEventListener("click", () => startShortcutCapture(action.id));

            const resetButton = document.createElement("button");
            resetButton.className = "ghost-button shortcut-button";
            resetButton.type = "button";
            resetButton.textContent = "重置";
            resetButton.addEventListener("click", () => {
                shortcutDraft[action.id] = DEFAULT_SHORTCUTS[action.id];
                capturingShortcutAction = "";
                renderShortcutRows();
            });

            row.append(title, input, recordButton, resetButton);
            shortcutList.append(row);
        }
    }

    function startShortcutCapture(actionId) {
        capturingShortcutAction = actionId;
        renderShortcutRows();
        showStatus("按下快捷键");
    }

    function captureShortcutFromKeyboard(event) {
        if (!capturingShortcutAction) {
            return;
        }

        event.preventDefault();
        event.stopPropagation();
        const shortcut = shortcutFromKeyboardEvent(event);
        if (!shortcut) {
            showStatus("请包含修饰键", true);
            return;
        }

        shortcutDraft[capturingShortcutAction] = shortcut;
        capturingShortcutAction = "";
        renderShortcutRows();
        showStatus("已录制");
    }

    function shortcutFromKeyboardEvent(event) {
        const key = normalizeKeyboardShortcutKey(event.key);
        if (!key) {
            return "";
        }

        const modifiers = [];
        if (event.metaKey) {
            modifiers.push(IS_MAC_PLATFORM ? "CmdOrCtrl" : "Super");
        }
        if (event.ctrlKey) {
            modifiers.push("Ctrl");
        }
        if (event.altKey) {
            modifiers.push("Alt");
        }
        if (event.shiftKey) {
            modifiers.push("Shift");
        }

        const isFunctionKey = /^F\d{1,2}$/.test(key);
        if (modifiers.length === 0 && !isFunctionKey) {
            return "";
        }
        return [...modifiers, key].join("+");
    }

    function normalizeKeyboardShortcutKey(key) {
        const aliases = {
            " ": "Space",
            Spacebar: "Space",
            Escape: "Escape",
            Esc: "Escape",
            ArrowLeft: "Left",
            ArrowRight: "Right",
            ArrowUp: "Up",
            ArrowDown: "Down",
            PageUp: "Page Up",
            PageDown: "Page Down",
            Delete: "Delete",
            Backspace: "Backspace",
            Enter: "Enter",
            Return: "Return",
            Tab: "Tab",
            Home: "Home",
            End: "End",
        };
        if (["Control", "Shift", "Alt", "Meta"].includes(key)) {
            return "";
        }
        if (aliases[key]) {
            return aliases[key];
        }
        const functionKeyMatch = /^F(\d{1,2})$/i.exec(key);
        if (functionKeyMatch) {
            const functionKeyNumber = Number(functionKeyMatch[1]);
            return functionKeyNumber >= 1 && functionKeyNumber <= 35 ? key.toUpperCase() : "";
        }
        if (key.length === 1) {
            return key.toUpperCase();
        }
        return "";
    }

    async function saveShortcuts() {
        shortcutSaveButton.disabled = true;
        try {
            settings = normalizeSettings(await callService("SetShortcuts", shortcutDraft));
            shortcutDraft = { ...settings.shortcuts };
            renderShortcutRows();
            showStatus("快捷键已应用");
        } catch (error) {
            showStatus("快捷键保存失败", true);
            console.error(error);
        } finally {
            shortcutSaveButton.disabled = false;
        }
    }

    async function savePetCount() {
        const nextCount = clampNumber(Number(numberInput.value), 1, settings.maxPets);
        saveButton.disabled = true;
        try {
            settings = normalizeSettings(await callService("SetPetCount", nextCount));
            applySettingsToControls(settings);
            renderPetImageSlots(settings);
            showStatus("已应用");
        } catch (error) {
            showStatus("保存失败", true);
            console.error(error);
        } finally {
            saveButton.disabled = false;
        }
    }

    function renderPetImageSlots(currentSettings) {
        imageList.replaceChildren();
        const images = readPetImages();
        const profiles = Object.values(PET_PROFILES);

        for (let slot = 0; slot < currentSettings.maxPets; slot += 1) {
            const profile = profiles[slot % profiles.length];
            const row = document.createElement("article");
            row.className = "pet-image-row";
            row.dataset.active = String(slot < currentSettings.petCount);

            const preview = document.createElement("div");
            preview.className = "pet-image-preview";

            const image = document.createElement("img");
            image.alt = "";
            image.src = images[String(slot)] || actionImageForSlot(slot);
            preview.append(image);

            const content = document.createElement("div");
            content.className = "pet-image-content";

            const title = document.createElement("h3");
            title.textContent = `${profile.label} ${slot + 1}`;

            const state = document.createElement("span");
            state.className = "pet-image-state";
            state.textContent = slot < currentSettings.petCount ? "已开启" : "未开启";

            const actions = document.createElement("div");
            actions.className = "pet-image-actions";

            const uploadLabel = document.createElement("label");
            uploadLabel.className = "secondary-button";
            uploadLabel.textContent = "上传";

            const input = document.createElement("input");
            input.type = "file";
            input.accept = "image/png,image/jpeg,image/webp,image/gif";
            input.addEventListener("change", async () => {
                const file = input.files?.[0];
                input.value = "";
                if (!file) {
                    return;
                }
                await uploadPetImage(slot, file);
            });
            uploadLabel.append(input);

            const resetButton = document.createElement("button");
            resetButton.className = "ghost-button";
            resetButton.type = "button";
            resetButton.textContent = "重置";
            resetButton.disabled = !images[String(slot)];
            resetButton.addEventListener("click", async () => {
                setStoredPetImage(slot, "");
                await notifyVisualsChanged();
                renderPetImageSlots(settings);
                showStatus("已重置");
            });

            actions.append(uploadLabel, resetButton);
            content.append(title, state, actions);
            row.append(preview, content);
            imageList.append(row);
        }
    }

    async function uploadPetImage(slot, file) {
        if (!file.type.startsWith("image/") || file.type === "image/svg+xml") {
            showStatus("请选择 PNG、JPG、WebP 或 GIF", true);
            return;
        }

        try {
            showStatus("正在处理图片");
            const dataUrl = await imageFileToPetDataUrl(file);
            setStoredPetImage(slot, dataUrl);
            await notifyVisualsChanged();
            renderPetImageSlots(settings);
            showStatus("已上传");
        } catch (error) {
            showStatus(error instanceof Error ? error.message : "上传失败", true);
            console.error(error);
        }
    }

    async function notifyVisualsChanged() {
        try {
            await callService("NotifyVisualSettingsChanged");
        } catch (error) {
            console.error(error);
        }
    }

    function showStatus(message, isError = false) {
        status.textContent = message;
        status.dataset.error = String(isError);
        window.clearTimeout(showStatus.timer);
        showStatus.timer = window.setTimeout(() => {
            status.textContent = "";
            status.dataset.error = "false";
        }, 2600);
    }
}

async function imageFileToPetDataUrl(file) {
    if (file.type === "image/gif") {
        if (file.size > CUSTOM_GIF_MAX_BYTES) {
            throw new Error("GIF 请控制在 4MB 内");
        }
        return readFileAsDataUrl(file);
    }

    const source = await readFileAsDataUrl(file);
    const image = await loadImage(source);
    const scale = Math.min(1, CUSTOM_IMAGE_MAX_SIDE / image.naturalWidth, CUSTOM_IMAGE_MAX_SIDE / image.naturalHeight);
    const width = Math.max(1, Math.round(image.naturalWidth * scale));
    const height = Math.max(1, Math.round(image.naturalHeight * scale));

    const canvas = document.createElement("canvas");
    canvas.width = width;
    canvas.height = height;
    const context = canvas.getContext("2d");
    context.clearRect(0, 0, width, height);
    context.drawImage(image, 0, 0, width, height);

    const dataUrl = canvas.toDataURL("image/webp", 0.9);
    if (dataUrl.length > 2.8 * 1024 * 1024) {
        throw new Error("图片过大，请换一张更小的图");
    }
    return dataUrl;
}

async function transparentActionImageUrl(src) {
    if (/\.(?:gif|webp)(?:$|[?#])/i.test(src)) {
        return src;
    }
    if (transparentActionImageUrls.has(src)) {
        return transparentActionImageUrls.get(src);
    }
    if (transparentActionImagePromises.has(src)) {
        return transparentActionImagePromises.get(src);
    }

    const promise = removeLightImageBackground(src)
        .then((url) => {
            transparentActionImageUrls.set(src, url);
            transparentActionImagePromises.delete(src);
            return url;
        })
        .catch((error) => {
            transparentActionImagePromises.delete(src);
            console.error(error);
            return src;
        });
    transparentActionImagePromises.set(src, promise);
    return promise;
}

async function removeLightImageBackground(src) {
    const image = await loadImage(src);
    const canvas = document.createElement("canvas");
    canvas.width = image.naturalWidth;
    canvas.height = image.naturalHeight;
    const context = canvas.getContext("2d");
    context.drawImage(image, 0, 0);

    const pixels = context.getImageData(0, 0, canvas.width, canvas.height);
    for (let index = 0; index < pixels.data.length; index += 4) {
        const red = pixels.data[index];
        const green = pixels.data[index + 1];
        const blue = pixels.data[index + 2];
        if (red >= BACKGROUND_ALPHA_THRESHOLD && green >= BACKGROUND_ALPHA_THRESHOLD && blue >= BACKGROUND_ALPHA_THRESHOLD) {
            pixels.data[index + 3] = 0;
        }
    }

    context.putImageData(pixels, 0, 0);
    return canvas.toDataURL("image/webp", 0.92);
}

function readFileAsDataUrl(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(String(reader.result || ""));
        reader.onerror = () => reject(reader.error || new Error("读取图片失败"));
        reader.readAsDataURL(file);
    });
}

function loadImage(src) {
    return new Promise((resolve, reject) => {
        const image = new Image();
        image.onload = () => resolve(image);
        image.onerror = () => reject(new Error("图片无法加载"));
        image.src = src;
    });
}
