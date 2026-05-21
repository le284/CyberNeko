import { Call, Events, Window } from "@wailsio/runtime";

const SERVICE_PREFIX = "main.PetService.";
const PET_IMAGES_STORAGE_KEY = "cyberneko.petImages.v1";
const CUSTOM_IMAGE_MAX_SIDE = 360;
const CUSTOM_GIF_MAX_BYTES = 4 * 1024 * 1024;

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
    kuro: Object.freeze({
        id: "kuro",
        label: "CyberKuro",
        menuId: "kuro-menu",
        greetingLine: "暗色模式也有我",
        doubleClickLine: "这下精神了",
        theme: Object.freeze({
            "--pet-fur": "#232833",
            "--pet-outline": "#090b10",
            "--pet-ear": "#67e8f9",
            "--pet-eye": "#f472b6",
            "--pet-eye-glow": "rgba(244, 114, 182, 0.72)",
            "--pet-bubble-text": "#181b23",
            "--pet-bubble-border": "rgba(9, 11, 16, 0.9)",
        }),
        speechLines: Object.freeze([
            "我会安静陪你",
            "边缘很适合散步",
            "让我追一下鼠标",
            "我把专注力递给你",
            "今天也别太累",
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
    const petCount = Number(rawSettings?.petCount ?? rawSettings?.PetCount ?? 2);
    const maxPets = Number(rawSettings?.maxPets ?? rawSettings?.MaxPets ?? 6);
    return {
        petCount: clampNumber(petCount, 1, maxPets || 6),
        maxPets: clampNumber(maxPets || 6, 1, 12),
    };
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
    const customImageElement = document.getElementById("pet-custom-image");
    const speechBubble = document.getElementById("neko-speech");

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
        pauseAutoMove(6000);
        manualDrag.active = true;
        manualDrag.ready = false;
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

    petElement.addEventListener("dblclick", () => {
        setPetState(currentState === PetState.Idle ? PetState.Walking : PetState.Idle);
        showSpeechBubble(petProfile.doubleClickLine);
    });

    setPetState(PetState.Idle);
    setPetDirection("right");
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
    const status = document.getElementById("settings-status");

    let settings = normalizeSettings(await safeGetSettings());
    applySettingsToControls(settings);
    renderPetImageSlots(settings);

    closeButton.addEventListener("click", () => Window.Hide());
    saveButton.addEventListener("click", savePetCount);
    rangeInput.addEventListener("input", () => syncCountInputs(rangeInput.value));
    numberInput.addEventListener("input", () => syncCountInputs(numberInput.value));

    Events.On("pet:settings", (event) => {
        settings = normalizeSettings(event.data);
        applySettingsToControls(settings);
        renderPetImageSlots(settings);
    });

    async function safeGetSettings() {
        try {
            return await callService("GetSettings");
        } catch (error) {
            showStatus("设置服务暂时不可用", true);
            console.error(error);
            return { petCount: 2, maxPets: 6 };
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
            image.src = images[String(slot)] || "/logo.svg";
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
